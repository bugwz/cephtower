package clusterinspect

import (
	"context"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestServiceDaemonsNativeQuery(t *testing.T) {
	s, runner, id := testInspection(t)
	runner.output = `[{"daemon_name":"rgw.a.host.x","service_name":"rgw.a","memory_usage":18446744073709551615,"events":["password=secret"]}]`
	result, err := s.ServiceDaemons(context.Background(), id, "rgw.a")
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(result)
	if !strings.Contains(string(raw), `"memory_usage":"18446744073709551615"`) {
		t.Fatal(string(raw))
	}
	if len(runner.specs) != 1 || runner.specs[0].Mutating || !reflect.DeepEqual(runner.specs[0].Args, []string{"orch", "ps", "--service-name", "rgw.a", "--refresh", "--format", "json"}) {
		t.Fatalf("%+v", runner.specs)
	}
	for _, output := range []string{`null`, `{}`, `[null]`, `[{"daemon_name":"a","service_name":"other"}]`, `[{"daemon_name":"a"},{"daemon_name":"a"}]`, `[{"daemon_name":"a","memory_usage":-1}]`, `[{"daemon_name":"a","memory_usage":1.5}]`} {
		runner.output = output
		if _, err := s.ServiceDaemons(context.Background(), id, "rgw.a"); err == nil {
			t.Fatalf("accepted %s", output)
		}
	}
	runner.output = `[]`
	if _, err := s.ServiceDaemons(context.Background(), id, "rgw.a"); err != nil {
		t.Fatal(err)
	}
	count := len(runner.specs)
	for _, name := range []string{"", "--help", "a b", "a/b"} {
		if _, err := s.ServiceDaemons(context.Background(), id, name); err == nil {
			t.Fatal(name)
		}
	}
	if len(runner.specs) != count {
		t.Fatal("invalid name executed")
	}
	runner.fail = true
	if _, err := s.ServiceDaemons(context.Background(), id, "rgw.a"); err == nil {
		t.Fatal("failure hidden")
	}
}
