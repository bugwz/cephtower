package mutation

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func rbdPeerIDs(data []byte) map[string]bool {
	var info struct {
		Mode  string `json:"mode"`
		Peers []struct {
			UUID string `json:"uuid"`
		} `json:"peers"`
	}
	if json.Unmarshal(data, &info) != nil || info.Peers == nil || (info.Mode != "image" && info.Mode != "pool" && info.Mode != "init-only" && info.Mode != "disabled") {
		return nil
	}
	seen := map[string]bool{}
	for _, peer := range info.Peers {
		id := strings.ToLower(peer.UUID)
		if id == "" || strings.TrimSpace(id) != id || seen[id] {
			return nil
		}
		seen[id] = true
	}
	return seen
}

func rbdPeerPresence(data []byte, target string, present bool) bool {
	seen := rbdPeerIDs(data)
	return seen != nil && seen[strings.ToLower(target)] == present
}

var rbdAddedPeerUUID = regexp.MustCompile(`(?i)^[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}$`)

func (s *Service) executeRBDPeerAddition(ctx context.Context, access executor.ClusterAccess, request Request, spec command) (cephdomain.ActionResult, error) {
	pool := optional(request.Parameters, "pool")
	read := func(suffix string) (executor.CommandResult, error) {
		return s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + suffix, Binary: executor.BinaryRBD, Args: []string{"mirror", "pool", "info", pool, "--format", "json"}, Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput})
	}
	before, err := read(".pre_check")
	ids := rbdPeerIDs(before.Stdout)
	if err != nil || ids == nil {
		return cephdomain.ActionResult{}, invalid("peer inventory could not be verified; refresh inventory before adding a peer")
	}
	written, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action, Binary: spec.binary, Args: spec.args, Timeout: spec.timeout, MaxOutput: executor.DefaultMaxOutput, Mutating: true})
	if err != nil {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "ceph_command_failed", Message: "peer addition outcome is uncertain; inspect the pool before any manual retry", Retryable: false}
	}
	// Native peer add prints only the new UUID and a newline. Do not infer an ID
	// from a list difference: another operator may have changed the pool.
	id := strings.TrimSuffix(string(written.Stdout), "\n")
	after, err := read(".post_check")
	matched := rbdAddedPeerUUID.MatchString(id) && !ids[strings.ToLower(id)] &&
		rbdPeerFieldMatches(after.Stdout, id, "site-name", optional(request.Parameters, "remote_cluster")) &&
		rbdPeerFieldMatches(after.Stdout, id, "client", optional(request.Parameters, "remote_client")) &&
		rbdPeerFieldMatches(after.Stdout, id, "direction", optional(request.Parameters, "direction"))
	if err != nil || !matched {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_check_failed", Message: "peer addition was not confirmed by readback; inspect the pool before any manual retry", Retryable: false}
	}
	return cephdomain.ActionResult{Details: map[string]any{"pool": pool, "peer_uuid": id, "created": true, "verified": true}}, nil
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
