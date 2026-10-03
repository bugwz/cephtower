package mutation

import (
	"context"
	"errors"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestDaemonActionUsesNativeScopedQuery(t *testing.T) {
	s, _, id := newCephUserService(t)
	for _, name := range []string{"osd.1", "rgw.realm.zone.node1.abcdef", "node-exporter.node1"} {
		for _, action := range []string{"start", "stop", "restart", "reconfig", "redeploy", "rotate-key"} {
			e := &directoryRenameExecutor{outputs: map[string]string{"daemon.action": "Scheduled to " + action + " " + name + " on host 'node1'"}}
			s.executor = e
			_, err := s.Execute(context.Background(), Request{ClusterID: id, Action: "daemon.action", ResourceKey: "daemon/" + name + "/action", Parameters: map[string]any{"action": action}})
			if err != nil || len(e.specs) != 2 {
				t.Fatalf("%s %s: %v %+v", name, action, err, e.specs)
			}
			var daemonType, daemonID string
			switch name {
			case "osd.1":
				daemonType, daemonID = "osd", "1"
			case "rgw.realm.zone.node1.abcdef":
				daemonType, daemonID = "rgw", "realm.zone.node1.abcdef"
			case "node-exporter.node1":
				daemonType, daemonID = "node-exporter", "node1"
			}
			if !e.specs[0].Mutating || !reflect.DeepEqual(e.specs[0].Args, []string{"orch", "daemon", action, name}) || e.specs[1].Mutating || !reflect.DeepEqual(e.specs[1].Args, []string{"orch", "ps", "--daemon-type", daemonType, "--daemon-id", daemonID, "--refresh", "--format", "json"}) {
				t.Fatalf("unexpected command chain: %+v", e.specs)
			}
		}
	}
}

func TestDaemonActionRejectsUnconfirmedScheduling(t *testing.T) {
	s, _, id := newCephUserService(t)
	for _, output := range []string{"", "success", "Scheduled to stop osd.1 on host 'node1'", "Scheduled to restart osd.2 on host 'node1'", "Scheduled to restart osd.1 on host ''", "Scheduled to restart osd.1 on host 'node1'\nfailed"} {
		e := &directoryRenameExecutor{outputs: map[string]string{"daemon.action": output}}
		s.executor = e
		_, err := s.Execute(context.Background(), Request{ClusterID: id, Action: "daemon.action", ResourceKey: "daemon/osd.1/action", Parameters: map[string]any{"action": "restart"}})
		var ae *cephdomain.ActionError
		if !errors.As(err, &ae) || ae.Code != "post_check_failed" || ae.Retryable || len(e.specs) != 1 {
			t.Fatalf("unconfirmed scheduling accepted: %q %v %+v", output, err, e.specs)
		}
	}
	if !daemonActionScheduled([]byte("Scheduled to restart osd.1 on host 'node1.example'\n"), "osd.1", "restart") {
		t.Fatal("native acknowledgement rejected")
	}
}

func TestDaemonActionRejectsMalformedIdentityBeforeExecution(t *testing.T) {
	s, _, id := newCephUserService(t)
	for _, name := range []string{"osd", "osd.", ".1", "--osd.1", "osd.--help", "osd.a b", "osd.a\nstop"} {
		e := &directoryRenameExecutor{}
		s.executor = e
		_, err := s.Execute(context.Background(), Request{ClusterID: id, Action: "daemon.action", ResourceKey: "daemon/" + name + "/action", Parameters: map[string]any{"action": "restart"}})
		if err == nil || len(e.specs) != 0 {
			t.Fatalf("malformed identity executed: %q %v %+v", name, err, e.specs)
		}
	}
}
