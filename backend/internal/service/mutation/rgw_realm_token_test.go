package mutation

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"cephtower/backend/internal/integration/ceph/executor"
)

type realmTokenExecutor struct {
	specs []executor.CommandSpec
	raw   string
	fail  bool
}

func (e *realmTokenExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	if e.fail {
		return executor.CommandResult{}, errors.New("must-not-expose-secret")
	}
	return executor.CommandResult{Stdout: []byte(e.raw)}, nil
}
func realmTokenFixture(t *testing.T, id, name string) string {
	t.Helper()
	b, err := json.Marshal(map[string]string{"realm_id": id, "realm_name": name, "endpoint": "https://master.example", "access_key": "fixture-access", "secret": "fixture-secret"})
	if err != nil {
		t.Fatal(err)
	}
	return base64.StdEncoding.EncodeToString(b)
}
func TestReadRGWRealmToken(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	token := realmTokenFixture(t, "id", "realm")
	raw, _ := json.Marshal([]map[string]string{{"realm": "other", "token": "realm has no master zone"}, {"realm": "realm", "token": token}})
	runner := &realmTokenExecutor{raw: string(raw)}
	s.executor = runner
	got, err := s.ReadRGWRealmToken(context.Background(), cluster, "id", "realm")
	if err != nil || got != token {
		t.Fatalf("unexpected token result: %v", err)
	}
	if len(runner.specs) != 1 || runner.specs[0].Mutating || runner.specs[0].Binary != executor.BinaryCeph || !reflect.DeepEqual(runner.specs[0].Args, []string{"rgw", "realm", "tokens"}) {
		t.Fatal("unexpected command")
	}
	for _, args := range [][2]string{{"", "realm"}, {"id", ""}, {"-id", "realm"}} {
		if _, err := s.ReadRGWRealmToken(context.Background(), cluster, args[0], args[1]); err == nil {
			t.Fatal("invalid identity accepted")
		}
	}
	if len(runner.specs) != 1 {
		t.Fatal("invalid requests executed commands")
	}
	runner.fail = true
	if _, err := s.ReadRGWRealmToken(context.Background(), cluster, "id", "realm"); err == nil || strings.Contains(err.Error(), "must-not-expose") {
		t.Fatal("command error leaked")
	}
}
func TestSelectRGWRealmTokenRejectsUnavailableAndMismatchedData(t *testing.T) {
	for _, token := range []string{"realm has no master zone", "master zone has no endpoint", "master zone has no access/secret keys", "bad", realmTokenFixture(t, "wrong", "realm"), realmTokenFixture(t, "id", "wrong"), base64.StdEncoding.EncodeToString([]byte(`{"realm_id":"id","realm_name":"realm"}`))} {
		raw, _ := json.Marshal([]map[string]string{{"realm": "realm", "token": token}})
		if _, err := selectRGWRealmToken(raw, "id", "realm"); err == nil || strings.Contains(err.Error(), token) {
			t.Fatal("invalid token accepted or exposed")
		}
	}
	for _, raw := range []string{"null", "{}", "[]", "[] {}", `[{"realm":"realm","token":"secret"},{"realm":"realm","token":"secret"}]`} {
		if _, err := selectRGWRealmToken([]byte(raw), "id", "realm"); err == nil {
			t.Fatal("invalid response accepted")
		}
	}
}
