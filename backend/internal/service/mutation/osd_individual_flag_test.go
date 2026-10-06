package mutation

import (
	"context"
	"reflect"
	"testing"
)

func TestOSDIndividualFlagExecution(t *testing.T) {
	for _, action := range []string{"set", "unset"} {
		for _, flag := range []string{"noout", "noin", "nodown", "noup"} {
			service, _, id := newCephUserService(t)
			state := `[]`
			if action == "set" {
				state = `["` + flag + `"]`
			}
			runner := &osdSafetyExecutor{output: `{"osds":[{"osd":0,"state":` + state + `}]}`}
			service.executor = runner
			_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "osd.individual_flag", ResourceKey: "osd/0", Parameters: map[string]any{"action": action, "flag": flag}})
			if err != nil {
				t.Fatal(err)
			}
			if len(runner.specs) != 2 || !runner.specs[0].Mutating || runner.specs[1].Mutating || !reflect.DeepEqual(runner.specs[0].Args, []string{"osd", action + "-group", flag, "0"}) || !reflect.DeepEqual(runner.specs[1].Args, []string{"osd", "dump", "--format", "json"}) {
				t.Fatalf("specs=%+v", runner.specs)
			}
		}
	}
}

func TestOSDIndividualFlagValidation(t *testing.T) {
	for _, tc := range []struct{ id, action, flag string }{{"all", "set", "noout"}, {"01", "set", "noout"}, {"0", "set-group", "noout"}, {"0", "set", "noscrub"}, {"0", "set", "noout,noin"}} {
		if _, err := build(Request{Action: "osd.individual_flag", ResourceKey: "osd/" + tc.id}, map[string]any{"action": tc.action, "flag": tc.flag}); err == nil {
			t.Fatalf("accepted %+v", tc)
		}
	}
	request := Request{ResourceKey: "osd/0", Parameters: map[string]any{"action": "unset", "flag": "noout"}}
	for _, state := range []string{`[null]`, `[""]`, `["up","up"]`, `[" noout"]`} {
		if osdIndividualFlagMatches(request, []byte(`{"osds":[{"osd":0,"state":`+state+`}]}`)) {
			t.Fatalf("accepted malformed state %s", state)
		}
	}
	for _, raw := range []string{`{}`, `{"osds":[]}`, `{"osds":[{"osd":1,"state":[]}]}`, `{"osds":[{"osd":0}]}`, `{"osds":[{"osd":0,"state":["noout"]}]}`, `{"osds":[{"osd":0,"state":[]},{"osd":0,"state":[]}]}`} {
		if osdIndividualFlagMatches(request, []byte(raw)) {
			t.Fatalf("accepted %s", raw)
		}
	}
}
