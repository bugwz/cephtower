package mutation

import (
	"context"
	"encoding/json"
	"slices"
	"strconv"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func osdDestroyState(raw []byte, id, uuid string, destroyed bool) bool {
	if destroyed {
		uuid = "00000000-0000-0000-0000-000000000000"
	}
	var dump struct {
		OSDs []struct {
			ID    *int     `json:"osd"`
			Up    *int     `json:"up"`
			UUID  *string  `json:"uuid"`
			State []string `json:"state"`
		} `json:"osds"`
	}
	if json.Unmarshal(raw, &dump) != nil || dump.OSDs == nil {
		return false
	}
	found := false
	for _, row := range dump.OSDs {
		if row.ID == nil {
			return false
		}
		if strconv.Itoa(*row.ID) != id {
			continue
		}
		if found || row.Up == nil || *row.Up != 0 || row.UUID == nil || *row.UUID != uuid || row.State == nil {
			return false
		}
		if slices.Contains(row.State, "destroyed") != destroyed {
			return false
		}
		seen := map[string]bool{}
		for _, state := range row.State {
			if state == "" || strings.TrimSpace(state) != state || seen[state] {
				return false
			}
			seen[state] = true
		}
		if len(row.State) == 0 {
			return false
		}
		found = true
	}
	return found
}

func (s *Service) executeOSDDestroy(ctx context.Context, access executor.ClusterAccess, request Request) (cephdomain.ActionResult, error) {
	id := last(resourceTail(request.ResourceKey))
	n, err := strconv.ParseUint(id, 10, 31)
	uuid, _ := request.Parameters["expected_uuid"].(string)
	confirmation, _ := request.Parameters["confirmation"].(string)
	if err != nil || strconv.FormatUint(n, 10) != id || uuid == "" || confirmation != "destroy osd."+id {
		return cephdomain.ActionResult{}, invalid("canonical OSD id, expected_uuid and exact destroy confirmation are required")
	}
	run := func(stage string, args []string, mutating bool) (executor.CommandResult, error) {
		result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: "osd.destroy." + stage, Binary: executor.BinaryCeph, Args: args, Mutating: mutating, Timeout: 45 * time.Second, MaxOutput: executor.DefaultMaxOutput})
		if err != nil || result.ExitCode != 0 {
			return result, &cephdomain.ActionError{Code: "ceph_command_failed", Message: "OSD destroy stage failed; inspect native OSD state before any manual retry", Retryable: false}
		}
		return result, nil
	}
	read := []string{"osd", "dump", "--format", "json"}
	before, err := run("pre_check", read, false)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	if !osdDestroyState(before.Stdout, id, uuid, false) {
		return cephdomain.ActionResult{}, invalid("OSD must be uniquely identified, down and not destroyed; refresh before proceeding")
	}
	safety, err := run("safety", []string{"osd", "safe-to-destroy", id, "--format", "json"}, false)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	report, valid := osdRemovalReport(safety.Stdout, []string{id})
	if !valid || report["is_safe_to_destroy"] != true {
		return cephdomain.ActionResult{}, invalid("native safety check did not confirm this OSD safe to destroy; no changes made")
	}
	if _, err := run("execute", []string{"osd", "destroy-actual", id, "--yes-i-really-mean-it"}, true); err != nil {
		return cephdomain.ActionResult{}, err
	}
	after, err := run("post_check", read, false)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	if !osdDestroyState(after.Stdout, id, uuid, true) {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_check_failed", Message: "destroy was issued but the destroyed state could not be verified; keys may already be removed; inspect before retrying", Retryable: false}
	}
	return cephdomain.ActionResult{Details: map[string]any{"osd_id": id, "destroyed": true}}, nil
}
