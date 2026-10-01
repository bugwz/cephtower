package mutation

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/base64"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestConfigurationDeletionConfirmsScopedAbsence(t *testing.T) {
	s, _, id := newCephUserService(t)
	r := Request{ClusterID: id, Action: "config_value.delete", ResourceKey: configurationTestKey("osd/class:ssd/host:node1", "test")}
	e := &directoryRenameExecutor{outputs: map[string]string{r.Action: "removed"}}
	s.executor = e
	for _, test := range []struct {
		data  string
		valid bool
	}{
		{`[]`, true},
		{`[{"section":"global","name":"test","value":""}]`, true},
		{`[{"section":"osd","mask":"class:ssd","name":"test","value":"0"}]`, true},
		{`[{"section":"osd","mask":"host:node1/class:ssd","name":"other","value":"false"}]`, true},
		{`[{"section":"osd","mask":"host:node1/class:ssd","name":"test","value":""}]`, false},
		{`[{"section":"osd","location_type":"host","location_value":"node1","device_class":"ssd","name":"test","value":"0"}]`, false},
		{`null`, false}, {`{}`, false}, {`[] {}`, false},
		{`[{"section":"global","name":"other"}]`, false},
		{`[{"section":"global","name":"other","value":null}]`, false},
		{`[{"section":"osd","location_type":"host","name":"other","value":""}]`, false},
		{`[{"section":"osd","mask":"class:hdd","device_class":"ssd","name":"other","value":""}]`, false},
		{`[{"section":"global","name":"other","value":""},{"section":"global","name":"other","value":""}]`, false},
	} {
		e.outputs[r.Action+".post_check"] = test.data
		_, err := s.Execute(context.Background(), r)
		if test.valid {
			if err != nil {
				t.Fatal(test.data, err)
			}
			continue
		}
		var failure *cephdomain.ActionError
		if !errors.As(err, &failure) || failure.Code != "post_check_failed" || failure.Retryable {
			t.Fatal(test.data, err)
		}
	}
	e.failID = r.Action + ".post_check"
	_, err := s.Execute(context.Background(), r)
	var failure *cephdomain.ActionError
	if !errors.As(err, &failure) || failure.Code != "post_check_failed" || failure.Retryable {
		t.Fatal(err)
	}
}

func TestConfigurationPreservesScopesAndValues(t *testing.T) {
	for _, value := range []string{"", "-1", "prefix with spaces", `{"setting":true}`, "--help", " x ", "first\nsecond", "--help\nsecond"} {
		cmd, err := build(Request{Action: "config_value.set", ResourceKey: configurationTestKey("mgr", "mgr/dashboard/ssl_server_port")}, map[string]any{"value": value})
		expected := []string{"config", "set", "mgr", "mgr/dashboard/ssl_server_port", "--value=" + value}
		if value == "" || strings.ContainsAny(value, "\r\n") {
			expected = []string{"--", "config", "set", "mgr", "mgr/dashboard/ssl_server_port", "--value", value}
		}
		if err != nil || !reflect.DeepEqual(cmd.args, expected) {
			t.Fatalf("value=%q command=%v err=%v", value, cmd.args, err)
		}
		if !reflect.DeepEqual(cmd.check, []string{"config", "dump", "--format", "json"}) {
			t.Fatalf("incorrect scoped post-check: %v", cmd.check)
		}
	}
	cmd, err := build(Request{Action: "config_value.delete", ResourceKey: configurationTestKey("osd/class:ssd", "osd_memory_target")}, nil)
	if err != nil || !reflect.DeepEqual(cmd.args, []string{"config", "rm", "osd/class:ssd", "osd_memory_target"}) {
		t.Fatalf("delete args=%v err=%v", cmd.args, err)
	}
	for _, resource := range []string{configurationTestKey("--help", "foo"), configurationTestKey("global", "--foo"), "configuration/value", configurationTestKey("osd;id", "foo")} {
		if _, err := build(Request{Action: "config_value.set", ResourceKey: resource}, map[string]any{"value": "x"}); err == nil {
			t.Fatalf("accepted %s", resource)
		}
	}
	for _, value := range []any{true, "x\x00", strings.Repeat("x", (32<<10)+1)} {
		if _, err := build(Request{Action: "config_value.set", ResourceKey: configurationTestKey("global", "test")}, map[string]any{"value": value}); err == nil {
			t.Fatal("invalid value accepted")
		}
	}
	cmd, err = build(Request{Action: "config_value.set", ResourceKey: configurationTestKey("client.rgw", "rgw_keystone_admin_password")}, map[string]any{"value": "test-secret"})
	if err != nil || len(cmd.sensitive) != 1 {
		t.Fatal("secret value not marked sensitive")
	}
}

func configurationTestKey(who, name string) string {
	return "configuration/value/" + base64.RawURLEncoding.EncodeToString([]byte(who+"\x00"+name))
}
