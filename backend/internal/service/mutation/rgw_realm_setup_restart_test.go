package mutation

import (
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"strings"
	"testing"
	"time"
)

func TestSetupRestartEvidence(t *testing.T) {
	r := setupResponses(t)
	before, ok := setupDaemonSnapshot([]byte(r["daemons_before"]), "rgw.gateway")
	if !ok {
		t.Fatal("invalid baseline")
	}
	after, ok := setupDaemonSnapshot([]byte(r["daemons_ready"]), "rgw.gateway")
	if !ok || !setupRestartReady(before, after) {
		t.Fatal("restart not recognized")
	}
	for _, pair := range [][2]string{
		{`"status":1`, `"status":0`},
		{`"status":1`, `"status":9`},
		{`"hostname":"host"`, `"hostname":"other"`},
		{`gateway.host.id`, `gateway.host.other`},
		{`00:02:00Z`, `00:00:00Z`},
		{`00:03:00Z`, `00:01:00Z`},
	} {
		changed, valid := setupDaemonSnapshot([]byte(strings.ReplaceAll(r["daemons_ready"], pair[0], pair[1])), "rgw.gateway")
		if !valid || setupRestartReady(before, changed) {
			t.Fatal("insufficient restart evidence accepted")
		}
	}
	for _, raw := range []string{`null`, `{}`, `[{}]`, strings.ReplaceAll(r["daemons_ready"], `rgw.gateway`, `rgw.other`), strings.ReplaceAll(r["daemons_ready"], `2026-01-01T00:02:00Z`, `invalid`)} {
		if _, valid := setupDaemonSnapshot([]byte(raw), "rgw.gateway"); valid {
			t.Fatal("invalid daemon response accepted")
		}
	}
	if setupRestartReady(before, before) || setupRestartReady(nil, nil) {
		t.Fatal("stale or absent evidence accepted")
	}
	calls := 0
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if !waitSetupRestart(ctx, time.Millisecond, before, func() (map[string]setupDaemon, bool) {
		calls++
		if calls == 1 {
			return map[string]setupDaemon{}, true
		}
		if calls == 2 {
			return before, true
		}
		return after, true
	}) || calls != 3 {
		t.Fatal("did not wait through empty and stale snapshots")
	}
	deadline, stop := context.WithTimeout(context.Background(), 5*time.Millisecond)
	defer stop()
	if waitSetupRestart(deadline, time.Millisecond, before, func() (map[string]setupDaemon, bool) { return before, true }) {
		t.Fatal("timeout reported success")
	}
	if waitSetupRestart(ctx, time.Millisecond, before, func() (map[string]setupDaemon, bool) { return nil, false }) {
		t.Fatal("read failure reported success")
	}
	canceled, stopCanceled := context.WithCancel(context.Background())
	stopCanceled()
	if waitSetupRestart(canceled, time.Millisecond, before, func() (map[string]setupDaemon, bool) { t.Fatal("read after cancellation"); return after, true }) {
		t.Fatal("cancellation reported success")
	}
}

func TestSetupRestartWaitsForTransientMetadata(t *testing.T) {
	r := setupResponses(t)
	before, _ := setupDaemonSnapshot([]byte(r["daemons_before"]), "rgw.gateway")
	ready, _ := setupDaemonSnapshot([]byte(r["daemons_ready"]), "rgw.gateway")
	for _, pair := range [][2]string{
		{`"status":1`, `"status":null`},
		{`"started":"2026-01-01T00:02:00Z"`, `"started":null`},
		{`"last_refresh":"2026-01-01T00:03:00Z"`, `"last_refresh":null`},
	} {
		pending, ok := setupDaemonSnapshot([]byte(strings.ReplaceAll(r["daemons_ready"], pair[0], pair[1])), "rgw.gateway")
		if !ok || setupRestartBaseline(pending) || setupRestartReady(before, pending) || setupRestartReady(pending, ready) {
			t.Fatal("transient metadata was rejected or accepted as restart evidence")
		}
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		calls := 0
		result := waitSetupRestart(ctx, time.Millisecond, before, func() (map[string]setupDaemon, bool) {
			calls++
			if calls == 1 {
				return pending, true
			}
			return ready, true
		})
		cancel()
		if !result || calls != 2 {
			t.Fatal("did not wait for complete metadata")
		}
	}
}

func TestSetupWithoutGatewaysDoesNotClaimReadiness(t *testing.T) {
	s, _, cluster := newCephUserService(t)
	p := setupParameters()
	p["expected_services"] = []string{}
	r := setupResponses(t)
	r["services_before"], r["services_check"] = `[]`, `[]`
	e := &realmSetupExecutor{responses: r}
	s.executor = e
	result, err := s.Execute(context.Background(), Request{ClusterID: cluster, Action: "rgw_realm.setup", Parameters: p})
	if err != nil || result.Details.(map[string]any)["daemons_verified"] != false {
		t.Fatal("empty gateway scope claimed readiness", err)
	}
	for _, spec := range e.specs {
		if spec.Binary == executor.BinaryCeph && strings.Contains(spec.ID, "daemon") {
			t.Fatal("queried an unconfirmed gateway")
		}
	}
}
