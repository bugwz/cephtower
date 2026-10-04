package mutation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

// ReadRGWRealmToken bypasses inventory and durable operations: the native result
// contains system credentials for every locally available master realm.
func (s *Service) ReadRGWRealmToken(ctx context.Context, clusterID uint64, realmID, name string) (string, error) {
	if clusterID == 0 || !syncFlowToken(realmID) || !syncFlowToken(name) {
		return "", invalid("cluster_id, realm_id and name are required")
	}
	access, err := s.clusters.Access(ctx, clusterID)
	if err != nil {
		return "", err
	}
	defer func() { access.ClientKey = "" }()
	result, err := s.executor.Run(ctx, access, executor.CommandSpec{
		ID: "rgw_realm.token.read", Binary: executor.BinaryCeph,
		Args: []string{"rgw", "realm", "tokens"}, Timeout: time.Minute, MaxOutput: 1 << 20,
	})
	defer func() { clear(result.Stdout); clear(result.Stderr) }()
	if err != nil {
		return "", &cephdomain.ActionError{Code: "ceph_command_failed", Message: "Realm token read failed; ensure the rgw manager module is available"}
	}
	return selectRGWRealmToken(result.Stdout, realmID, name)
}

func selectRGWRealmToken(raw []byte, realmID, name string) (string, error) {
	failure := func() (string, error) {
		return "", &cephdomain.ActionError{Code: "capability_unavailable", Message: "Realm token unavailable or identity mismatch; check the local master zone, endpoint and system credentials"}
	}
	var entries []struct {
		Realm string `json:"realm"`
		Token string `json:"token"`
	}
	if len(raw) > 1<<20 || json.Unmarshal(raw, &entries) != nil || entries == nil {
		return failure()
	}
	seen := map[string]bool{}
	selected := ""
	for _, entry := range entries {
		if entry.Realm == "" || seen[entry.Realm] {
			return failure()
		}
		seen[entry.Realm] = true
		if entry.Realm == name {
			selected = entry.Token
		}
	}
	if selected == "" || len(selected) > 64<<10 {
		return failure()
	}
	decoded, err := base64.StdEncoding.Strict().DecodeString(selected)
	if err != nil {
		return failure()
	}
	defer clear(decoded)
	if base64.StdEncoding.EncodeToString(decoded) != selected {
		return failure()
	}
	var token struct {
		RealmID   string `json:"realm_id"`
		RealmName string `json:"realm_name"`
		Endpoint  string `json:"endpoint"`
		AccessKey string `json:"access_key"`
		Secret    string `json:"secret"`
	}
	if json.Unmarshal(decoded, &token) != nil || token.RealmID != realmID || token.RealmName != name || token.Endpoint == "" || token.AccessKey == "" || token.Secret == "" {
		return failure()
	}
	return selected, nil
}
