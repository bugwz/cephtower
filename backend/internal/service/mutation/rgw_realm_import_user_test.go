package mutation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

func TestImportedSystemUserEvidence(t *testing.T) {
	token := rgwRealmTokenDocument{AccessKey: "fixture-access", Secret: "fixture-secret"}
	raw := realmImportResponses()["system_user"]
	if !realmImportedSystemUser([]byte(raw), token) {
		t.Fatal("valid system user rejected")
	}
	for _, bad := range []string{`null`, `{}`, strings.ReplaceAll(raw, `"system":true`, `"system":false`), strings.ReplaceAll(raw, `"user_id":"sys"`, `"user_id":""`), strings.ReplaceAll(raw, `fixture-secret`, `other-secret`), strings.ReplaceAll(raw, `fixture-access`, `other-access`), `{"user_id":"sys","system":true,"keys":[{"access_key":"fixture-access","secret_key":"fixture-secret"},{"access_key":"fixture-access","secret_key":"fixture-secret"}]}`} {
		if realmImportedSystemUser([]byte(bad), token) {
			t.Fatal("invalid system user accepted")
		}
	}
	withOtherKey := strings.ReplaceAll(raw, `"keys":[`, `"keys":[{"access_key":"other","secret_key":"another"},`)
	if !realmImportedSystemUser([]byte(withOtherKey), token) {
		t.Fatal("unrelated key prevented verification")
	}
	s, _, cluster := newCephUserService(t)
	e := &realmImportExecutor{responses: realmImportResponses()}
	e.responses["system_user"] = strings.ReplaceAll(raw, `fixture-secret`, `other-secret`)
	s.executor = e
	result, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.import", Parameters: realmImportParameters(t)})
	if err == nil || strings.Contains(err.Error(), "secret") {
		t.Fatal("unsafe verification failure")
	}
	encoded, _ := json.Marshal(result)
	if strings.Contains(string(encoded), "fixture") {
		t.Fatal("credentials leaked")
	}
}

type missingImportUserExecutor struct {
	cancel context.CancelFunc
	calls  int
	output []byte
}

func (e *missingImportUserExecutor) Run(_ context.Context, _ executor.ClusterAccess, _ executor.CommandSpec) (executor.CommandResult, error) {
	e.calls++
	e.cancel()
	return executor.CommandResult{ExitCode: 22, Stderr: e.output}, nil
}

func TestImportedSystemUserMissingLookupRespectsCancellation(t *testing.T) {
	s, _, _ := newCephUserService(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	e := &missingImportUserExecutor{cancel: cancel, output: []byte("lookup-error")}
	s.executor = e
	if s.verifyImportedSystemUser(ctx, executor.ClusterAccess{}, Request{Action: "rgw_realm.import"}, "zone-id", rgwRealmTokenDocument{AccessKey: "key", Secret: "secret"}) || e.calls != 1 {
		t.Fatal("lookup ignored cancellation")
	}
	for _, b := range e.output {
		if b != 0 {
			t.Fatal("raw error retained")
		}
	}
}
