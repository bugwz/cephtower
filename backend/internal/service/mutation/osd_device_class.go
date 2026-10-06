package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"time"
)

func osdDeviceClass(raw []byte, id int) (string, bool) {
	var rows []struct {
		ID    *int    `json:"osd"`
		Class *string `json:"device_class"`
	}
	if json.Unmarshal(raw, &rows) != nil || len(rows) != 1 || rows[0].ID == nil || *rows[0].ID != id || rows[0].Class == nil {
		return "", false
	}
	return *rows[0].Class, true
}

func (s *Service) executeOSDDeviceClass(ctx context.Context, access executor.ClusterAccess, request Request) (cephdomain.ActionResult, error) {
	id := last(resourceTail(request.ResourceKey))
	n, err := strconv.ParseUint(id, 10, 31)
	if err != nil || strconv.FormatUint(n, 10) != id {
		return cephdomain.ActionResult{}, invalid("invalid OSD id")
	}
	class, ok := request.Parameters["device_class"].(string)
	old, oldOK := request.Parameters["expected_class"].(string)
	if !ok || !oldOK || !regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,127}$`).MatchString(class) {
		return cephdomain.ActionResult{}, invalid("device_class and expected_class are required")
	}
	run := func(stage string, args []string, mutating bool) (executor.CommandResult, error) {
		result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + "." + stage, Binary: executor.BinaryCeph, Args: args, Mutating: mutating, Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput})
		if err != nil || result.ExitCode != 0 {
			return result, &cephdomain.ActionError{Code: "ceph_command_failed", Message: "device class operation was not confirmed; inspect current class before any manual retry", Retryable: false}
		}
		return result, nil
	}
	readArgs := []string{"osd", "crush", "get-device-class", id, "--format", "json"}
	before, err := run("pre_check", readArgs, false)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	current, valid := osdDeviceClass(before.Stdout, int(n))
	if !valid || current != old {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "resource_conflict", Message: "native OSD device class differs from the expected class; refresh before changing it", Retryable: false}
	}
	if current == class {
		return cephdomain.ActionResult{Details: map[string]any{"device_class": class, "changed": false}}, nil
	}
	if current != "" {
		if _, err := run("remove", []string{"osd", "crush", "rm-device-class", id}, true); err != nil {
			return cephdomain.ActionResult{}, err
		}
	}
	if _, err := run("set", []string{"osd", "crush", "set-device-class", class, id}, true); err != nil {
		return cephdomain.ActionResult{}, err
	}
	after, err := run("post_check", readArgs, false)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	actual, valid := osdDeviceClass(after.Stdout, int(n))
	if !valid || actual != class {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_check_failed", Message: "device class change may be partial; target class could not be verified; inspect before retrying", Retryable: false}
	}
	return cephdomain.ActionResult{Details: map[string]any{"device_class": actual, "changed": true}}, nil
}
