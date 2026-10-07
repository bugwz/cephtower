package mutation

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func rbdPeerPresence(data []byte, target string, present bool) bool {
	var info struct {
		Mode  string `json:"mode"`
		Peers []struct {
			UUID string `json:"uuid"`
		} `json:"peers"`
	}
	if json.Unmarshal(data, &info) != nil || info.Peers == nil || (info.Mode != "image" && info.Mode != "pool" && info.Mode != "init-only" && info.Mode != "disabled") {
		return false
	}
	seen := map[string]bool{}
	for _, peer := range info.Peers {
		id := strings.ToLower(peer.UUID)
		if id == "" || strings.TrimSpace(id) != id || seen[id] {
			return false
		}
		seen[id] = true
	}
	return seen[strings.ToLower(target)] == present
}

func rbdPeerFieldMatches(data []byte, target, field, value string) bool {
	if !rbdPeerPresence(data, target, true) {
		return false
	}
	key := map[string]string{"site-name": "site_name", "client": "client_name", "direction": "direction", "mon-host": "mon_host"}[field]
	if key == "" {
		return false
	}
	var info struct {
		Peers []map[string]json.RawMessage `json:"peers"`
	}
	if json.Unmarshal(data, &info) != nil {
		return false
	}
	for _, peer := range info.Peers {
		var id, actual string
		if json.Unmarshal(peer["uuid"], &id) == nil && strings.EqualFold(id, target) {
			return json.Unmarshal(peer[key], &actual) == nil && actual == value
		}
	}
	return false
}

func (s *Service) executeRBDPeerMutation(ctx context.Context, access executor.ClusterAccess, request Request, spec command) (cephdomain.ActionResult, error) {
	pool, id := optional(request.Parameters, "pool"), optional(request.Parameters, "uuid")
	remove := optional(request.Parameters, "action") == "remove"
	field := optional(request.Parameters, "field")
	read := func(suffix string) (executor.CommandResult, error) {
		args := []string{"mirror", "pool", "info", pool, "--format", "json"}
		if !remove && field == "mon-host" && suffix == ".post_check" {
			args = append(args, "--all")
		}
		return s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + suffix, Binary: executor.BinaryRBD, Args: args, Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput})
	}
	before, err := read(".pre_check")
	if err != nil || !rbdPeerPresence(before.Stdout, id, true) {
		return cephdomain.ActionResult{}, invalid("peer existence could not be uniquely verified in the requested pool; refresh inventory before changing it")
	}
	if _, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action, Binary: spec.binary, Args: spec.args, Timeout: spec.timeout, MaxOutput: executor.DefaultMaxOutput, Mutating: true}); err != nil {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "ceph_command_failed", Message: "peer mutation outcome is uncertain; inspect the pool before any manual retry", Retryable: false}
	}
	after, err := read(".post_check")
	// --all can contain peer credentials. Never return its output or underlying errors.
	defer clear(after.Stdout)
	defer clear(after.Stderr)
	matched := remove && rbdPeerPresence(after.Stdout, id, false) || !remove && rbdPeerFieldMatches(after.Stdout, id, field, optional(request.Parameters, "value"))
	if err != nil || !matched {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_check_failed", Message: "peer mutation was not confirmed by readback; inspect the pool before any manual retry", Retryable: false}
	}
	details := map[string]any{"pool": pool, "peer_uuid": id, "verified": true}
	if remove {
		details["deleted"] = true
	} else {
		details["field"] = field
	}
	return cephdomain.ActionResult{Details: details}, nil
}
