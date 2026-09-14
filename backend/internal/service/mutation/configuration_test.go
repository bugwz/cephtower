package mutation

import (
	"encoding/base64"
	"reflect"
	"strings"
	"testing"
)

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
