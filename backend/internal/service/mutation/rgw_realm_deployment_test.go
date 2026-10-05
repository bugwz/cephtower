package mutation

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestRealmDeploymentEvidence(t *testing.T) {
	r := realmImportResponses()
	var expected []map[string]any
	if json.Unmarshal([]byte(r["service_post_check"]), &expected) != nil {
		t.Fatal("bad fixture")
	}
	check := func(service, daemon string, wantReady, wantValid bool) {
		t.Helper()
		ready, valid := realmDeploymentReady([]byte(service), []byte(daemon), expected[0])
		if ready != wantReady || valid != wantValid {
			t.Fatalf("got %v %v; want %v %v", ready, valid, wantReady, wantValid)
		}
	}
	s, d := r["deployment_service"], r["deployment_daemons"]
	check(s, d, true, true)
	for _, empty := range []string{`[]`, `No services reported`} {
		check(empty, `[]`, false, true)
	}
	check(s, `[]`, false, true)
	check(strings.ReplaceAll(s, `"running":1`, `"running":0`), d, false, true)
	check(strings.ReplaceAll(s, `"size":1`, `"size":2`), d, false, true)
	check(strings.ReplaceAll(s, `"size":1`, `"size":1.5`), d, false, false)
	check(strings.ReplaceAll(s, `"rgw_zone":"secondary"`, `"rgw_zone":"other"`), d, false, false)
	check(strings.ReplaceAll(s, `"status":{`, `"unmanaged":true,"status":{`), d, false, false)
	check(s, strings.ReplaceAll(d, `"status":1`, `"status":2`), false, true)
	check(s, strings.ReplaceAll(d, `"started":"2026-01-01T00:02:00Z"`, `"started":null`), false, true)
	check(s, strings.ReplaceAll(d, `"service_name":"rgw.realm.secondary"`, `"service_name":"rgw.other"`), false, false)
	check(s, `null`, false, false)
	expected[0]["spec"].(map[string]any)["update_endpoints"] = true
	check(strings.ReplaceAll(s, `"rgw_frontend_port":80`, `"rgw_frontend_port":80,"update_endpoints":false`), d, true, true)
	delete(expected[0]["spec"].(map[string]any), "update_endpoints")
	expected[0]["placement"] = map[string]any{"count": float64(2)}
	check(strings.ReplaceAll(s, `"status":{`, `"placement":{"count":2},"status":{`), d, false, true)
}

func TestRealmDeploymentWaitBounds(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	calls := 0
	if !waitRealmDeployment(ctx, time.Millisecond, func() (bool, bool) { calls++; return calls == 3, true }) || calls != 3 {
		t.Fatal("did not wait for deployment")
	}
	if waitRealmDeployment(ctx, time.Millisecond, func() (bool, bool) { return false, false }) {
		t.Fatal("invalid response accepted")
	}
	deadline, stop := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer stop()
	if waitRealmDeployment(deadline, time.Millisecond, func() (bool, bool) { return false, true }) {
		t.Fatal("timeout accepted")
	}
	stop()
	if waitRealmDeployment(deadline, time.Millisecond, func() (bool, bool) { t.Fatal("read after cancellation"); return true, true }) {
		t.Fatal("cancellation accepted")
	}
}
