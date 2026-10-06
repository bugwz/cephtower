package mutation

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func osdLostEpochs(raw []byte, id string) (down, lost uint32, valid bool) {
	var dump struct {
		OSDs []struct {
			ID   *int    `json:"osd"`
			Down *uint32 `json:"down_at"`
			Lost *uint32 `json:"lost_at"`
		} `json:"osds"`
	}
	if json.Unmarshal(raw, &dump) != nil || dump.OSDs == nil {
		return
	}
	for _, row := range dump.OSDs {
		if row.ID == nil {
			return 0, 0, false
		}
		if strconv.Itoa(*row.ID) != id {
			continue
		}
		if valid || row.Down == nil || row.Lost == nil || *row.Down == 0 {
			return 0, 0, false
		}
		down, lost, valid = *row.Down, *row.Lost, true
	}
	return
}

func (s *Service) executeOSDLost(ctx context.Context, access executor.ClusterAccess, request Request) (cephdomain.ActionResult, error) {
	id := last(resourceTail(request.ResourceKey))
	n, err := strconv.ParseUint(id, 10, 31)
	uuid, _ := request.Parameters["expected_uuid"].(string)
	confirmation, _ := request.Parameters["confirmation"].(string)
	if err != nil || strconv.FormatUint(n, 10) != id || uuid == "" || confirmation != "mark lost osd."+id {
		return cephdomain.ActionResult{}, invalid("canonical OSD id, expected_uuid and exact lost confirmation are required")
	}
	run := func(stage string, args []string, mutating bool) (executor.CommandResult, error) {
		result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: "osd.lost." + stage, Binary: executor.BinaryCeph, Args: args, Mutating: mutating, Timeout: 45 * time.Second, MaxOutput: executor.DefaultMaxOutput})
		if err != nil || result.ExitCode != 0 {
			return result, &cephdomain.ActionError{Code: "ceph_command_failed", Message: "OSD lost stage failed; permanent data loss is possible; inspect before any manual retry", Retryable: false}
		}
		return result, nil
	}
	read := []string{"osd", "dump", "--format", "json"}
	before, err := run("pre_check", read, false)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	down, _, valid := osdLostEpochs(before.Stdout, id)
	if !valid || !osdDestroyState(before.Stdout, id, uuid, false) {
		return cephdomain.ActionResult{}, invalid("OSD identity, down state and native epochs must be known; refresh before proceeding")
	}
	safety, err := run("safety", []string{"osd", "safe-to-destroy", id, "--format", "json"}, false)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	report, valid := osdRemovalReport(safety.Stdout, []string{id})
	if !valid || report["is_safe_to_destroy"] != true {
		return cephdomain.ActionResult{}, invalid("native safety check did not confirm this OSD safe to mark lost; no changes made")
	}
	if _, err := run("execute", []string{"osd", "lost", id, "--yes-i-really-mean-it"}, true); err != nil {
		return cephdomain.ActionResult{}, err
	}
	after, err := run("post_check", read, false)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	afterDown, lost, valid := osdLostEpochs(after.Stdout, id)
	if !valid || afterDown != down || lost != down || !osdDestroyState(after.Stdout, id, uuid, false) {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_check_failed", Message: "lost was issued but native lost_at/down_at state could not be verified; inspect before retrying", Retryable: false}
	}
	return cephdomain.ActionResult{Details: map[string]any{"osd_id": id, "lost_at": lost, "down_at": afterDown}}, nil
}
