package mutation

import (
	"context"
	"reflect"
	"testing"
)

func TestDaemonActionUsesNativeScopedQuery(t *testing.T) {
	s, _, id := newCephUserService(t)
	for _, name := range []string{"osd.1", "rgw.realm.zone.node1.abcdef", "node-exporter.node1"} {
		for _, action := range []string{"start", "stop", "restart", "reconfig", "redeploy", "rotate-key"} {
			e := &directoryRenameExecutor{}
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
