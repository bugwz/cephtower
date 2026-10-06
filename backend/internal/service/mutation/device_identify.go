package mutation

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

// The CLI lights every location of a device ID, unlike Dashboard's host/path API.
// Only accept a single native location matching the selected inventory target.
func deviceLightTarget(raw []byte, id, host, path string) bool {
	var device struct {
		ID        string `json:"devid"`
		Locations []struct {
			Host string `json:"host"`
			Dev  string `json:"dev"`
			Path string `json:"path"`
		} `json:"location"`
	}
	if json.Unmarshal(raw, &device) != nil || device.ID != id || len(device.Locations) != 1 {
		return false
	}
	loc := device.Locations[0]
	return loc.Host == host && (loc.Path == path || (loc.Dev != "" && !strings.Contains(loc.Dev, "/") && "/dev/"+loc.Dev == path))
}

func (s *Service) executeDeviceIdentify(ctx context.Context, access executor.ClusterAccess, request Request) (cephdomain.ActionResult, error) {
	p := request.Parameters
	id, err := required(p, "device_id")
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	host, err := required(p, "host")
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	path, err := required(p, "device")
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	if p["device_id"] != id || p["host"] != host || p["device"] != path || strings.HasPrefix(id, "-") || !strings.HasPrefix(path, "/dev/") {
		return cephdomain.ActionResult{}, invalid("device identity must be exact and include an absolute device path")
	}
	state, err := enum(p, "state", "on", "off")
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	light := optional(p, "light")
	if light == "" {
		light = "ident"
	}
	if light != "ident" && light != "fault" {
		return cephdomain.ActionResult{}, invalid("light is not supported")
	}
	run := func(stage string, args []string, mutating bool) (executor.CommandResult, error) {
		result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + "." + stage, Binary: executor.BinaryCeph, Args: args, Mutating: mutating, Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput})
		if err != nil || result.ExitCode != 0 {
			return result, &cephdomain.ActionError{Code: "ceph_command_failed", Message: "device light operation was not confirmed; inspect the physical device before retrying", Retryable: false}
		}
		return result, nil
	}
	before, err := run("pre_check", []string{"device", "info", id, "--format", "json"}, false)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	if !deviceLightTarget(before.Stdout, id, host, path) {
		return cephdomain.ActionResult{}, invalid("device ID must resolve to exactly the selected host and device location; no changes were made")
	}
	if _, err = run("apply", []string{"device", "light", state, id, light}, true); err != nil {
		return cephdomain.ActionResult{}, err
	}
	return cephdomain.ActionResult{Details: map[string]any{"device_id": id, "host": host, "device": path, "state": state, "light": light, "physical_state_verified": false}}, nil
}
