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

func (s *Service) executeRBDPeerRemoval(ctx context.Context, access executor.ClusterAccess, request Request, spec command) (cephdomain.ActionResult, error) {
	pool, id := optional(request.Parameters, "pool"), optional(request.Parameters, "uuid")
	read := func(suffix string) (executor.CommandResult, error) {
		return s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action + suffix, Binary: executor.BinaryRBD, Args: []string{"mirror", "pool", "info", pool, "--format", "json"}, Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput})
	}
	before, err := read(".pre_check")
	if err != nil || !rbdPeerPresence(before.Stdout, id, true) {
		return cephdomain.ActionResult{}, invalid("peer existence could not be uniquely verified in the requested pool; refresh inventory before removal")
	}
	if _, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: request.Action, Binary: spec.binary, Args: spec.args, Timeout: spec.timeout, MaxOutput: executor.DefaultMaxOutput, Mutating: true}); err != nil {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "ceph_command_failed", Message: "peer removal outcome is uncertain; inspect the pool before any manual retry", Retryable: false}
	}
	after, err := read(".post_check")
	if err != nil || !rbdPeerPresence(after.Stdout, id, false) {
		return cephdomain.ActionResult{}, &cephdomain.ActionError{Code: "post_check_failed", Message: "peer removal was not confirmed by readback; inspect the pool before any manual retry", Retryable: false}
	}
	return cephdomain.ActionResult{Details: map[string]any{"pool": pool, "peer_uuid": id, "deleted": true, "verified": true}}, nil
}
