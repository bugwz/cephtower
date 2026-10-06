package mutation

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func osdPurgeAbsent(raw []byte, array, field, id string) bool {
	var document map[string]json.RawMessage
	if json.Unmarshal(raw, &document) != nil {
		return false
	}
	var rows []map[string]json.RawMessage
	if json.Unmarshal(document[array], &rows) != nil || rows == nil {
		return false
	}
	seen := map[int32]bool{}
	for _, row := range rows {
		var value *int32
		if json.Unmarshal(row[field], &value) != nil || value == nil || *value < 0 || seen[*value] || strconv.FormatInt(int64(*value), 10) == id {
			return false
		}
		seen[*value] = true
	}
	return true
}

func osdPurgeCrushRemoved(raw []byte, id string) bool {
	var dump struct {
		Devices []struct {
			ID   *int32  `json:"id"`
			Name *string `json:"name"`
		} `json:"devices"`
		Buckets []struct {
			Items []struct {
				ID *int32 `json:"id"`
			} `json:"items"`
		} `json:"buckets"`
	}
	if json.Unmarshal(raw, &dump) != nil || dump.Devices == nil || dump.Buckets == nil {
		return false
	}
	seen := map[int32]bool{}
	for _, device := range dump.Devices {
		if device.ID == nil || *device.ID < 0 || device.Name == nil || seen[*device.ID] {
			return false
		}
		seen[*device.ID] = true
		if *device.Name == "osd."+id {
			return false
		}
		if strconv.FormatInt(int64(*device.ID), 10) == id && *device.Name != "device"+id {
			return false
		}
	}
	for _, bucket := range dump.Buckets {
		if bucket.Items == nil {
			return false
		}
		for _, item := range bucket.Items {
			if item.ID == nil || strconv.FormatInt(int64(*item.ID), 10) == id {
				return false
			}
		}
	}
	return true
}

func (s *Service) executeOSDPurge(ctx context.Context, access executor.ClusterAccess, request Request) (cephdomain.ActionResult, error) {
	id := last(resourceTail(request.ResourceKey))
	n, err := strconv.ParseUint(id, 10, 31)
	uuid, _ := request.Parameters["expected_uuid"].(string)
	confirmation, _ := request.Parameters["confirmation"].(string)
	if err != nil || strconv.FormatUint(n, 10) != id || uuid == "" || confirmation != "purge osd."+id {
		return cephdomain.ActionResult{}, invalid("canonical OSD id, expected_uuid and exact purge confirmation are required")
	}
	run := func(stage string, args []string, mutating bool) (executor.CommandResult, error) {
		result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: "osd.purge." + stage, Binary: executor.BinaryCeph, Args: args, Mutating: mutating, Timeout: 45 * time.Second, MaxOutput: executor.DefaultMaxOutput})
		if err != nil || result.ExitCode != 0 {
			return result, &cephdomain.ActionError{Code: "ceph_command_failed", Message: "OSD purge stage failed; identity, keys and CRUSH entry may have been removed; inspect before any manual retry", Retryable: false}
		}
		return result, nil
	}
	read := []string{"osd", "dump", "--format", "json"}
	before, err := run("pre_check", read, false)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	destroyed := uuid == "00000000-0000-0000-0000-000000000000"
	if !osdDestroyState(before.Stdout, id, uuid, destroyed) {
		return cephdomain.ActionResult{}, invalid("OSD must be uniquely identified and down; refresh before proceeding")
	}
	safety, err := run("safety", []string{"osd", "safe-to-destroy", id, "--format", "json"}, false)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	report, valid := osdRemovalReport(safety.Stdout, []string{id})
	if !valid || report["is_safe_to_destroy"] != true {
		return cephdomain.ActionResult{}, invalid("native safety check did not confirm this OSD safe to purge; no changes made")
	}
	if _, err := run("execute", []string{"osd", "purge-actual", id, "--yes-i-really-mean-it"}, true); err != nil {
		return cephdomain.ActionResult{}, err
	}
	after, err := run("post_check", read, false)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	if !osdPurgeAbsent(after.Stdout, "osds", "osd", id) {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_check_failed", Message: "purge was issued but OSD map removal could not be verified; inspect before retrying", Retryable: false}
	}
	crush, err := run("crush_check", []string{"osd", "crush", "dump", "--format", "json"}, false)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	if !osdPurgeCrushRemoved(crush.Stdout, id) {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_check_failed", Message: "purge was issued but CRUSH removal could not be verified; inspect before retrying", Retryable: false}
	}
	return cephdomain.ActionResult{Details: map[string]any{"osd_id": id, "purged": true}}, nil
}
