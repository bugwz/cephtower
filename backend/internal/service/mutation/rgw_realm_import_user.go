package mutation

import (
	"context"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

func realmImportedSystemUser(raw []byte, token rgwRealmTokenDocument) bool {
	user := periodDocument(raw)
	if user == nil || user["system"] != true || rawText(user, "user_id") == "" {
		return false
	}
	keys, ok := user["keys"].([]any)
	if !ok {
		return false
	}
	matches := 0
	for _, value := range keys {
		key, ok := value.(map[string]any)
		if !ok {
			return false
		}
		if key["access_key"] == token.AccessKey {
			if key["secret_key"] != token.Secret {
				return false
			}
			matches++
		}
	}
	return matches == 1
}

func (s *Service) verifyImportedSystemUser(ctx context.Context, access executor.ClusterAccess, request Request, zoneID string, token rgwRealmTokenDocument) bool {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	return waitRealmDeployment(ctx, 5*time.Second, func() (bool, bool) {
		result, err := s.executor.Run(ctx, access, executor.CommandSpec{
			ID: request.Action + ".system_user", Binary: executor.BinaryRGWAdmin,
			Args:          []string{"user", "info", "--access-key", token.AccessKey, "--zone-id", zoneID, "--format", "json"},
			SensitiveArgs: map[int]struct{}{3: {}}, Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput,
		})
		defer clear(result.Stdout)
		defer clear(result.Stderr)
		// RGWUser::info returns EINVAL when access-key lookup has not populated
		// the user. This is not proof of absence: wait only within the deadline.
		if result.ExitCode == 22 {
			return false, true
		}
		if err != nil || result.ExitCode != 0 {
			return false, false
		}
		verified := realmImportedSystemUser(result.Stdout, token)
		return verified, verified
	})
}
