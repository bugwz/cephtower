package mutation

import (
	"context"
	"testing"
)

func TestOSDGlobalFlagReadback(t *testing.T) {
	for _, tc := range []struct {
		action, flag, raw string
		valid             bool
	}{
		{"set", "noout", `{"flags":"noout,sortbitwise"}`, true},
		{"unset", "noout", `{"flags":"sortbitwise"}`, true},
		{"unset", "noout", `{"flags":""}`, true},
		{"set", "noout", `{"flags":""}`, false},
		{"unset", "noout", `{"flags":"noout"}`, false},
		{"unset", "noout", `{}`, false},
		{"unset", "noout", `{"flags":null}`, false},
		{"unset", "noout", `{"flags":[]}`, false},
		{"unset", "noout", `{"flags":"up,up"}`, false},
		{"unset", "noout", `{"flags":" noout"}`, false},
		{"set", "pause", `{"flags":"pauserd,pausewr"}`, true},
		{"set", "pause", `{"flags":"pauserd"}`, false},
		{"unset", "pause", `{"flags":"pausewr"}`, false},
		{"unset", "pause", `{"flags":""}`, true},
	} {
		parameters := map[string]any{"action": tc.action, "flag": tc.flag}
		if got := osdGlobalFlagMatches(parameters, []byte(tc.raw)); got != tc.valid {
			t.Fatalf("%+v got=%v", tc, got)
		}
		service, _, id := newCephUserService(t)
		service.executor = &osdSafetyExecutor{output: tc.raw}
		_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "osd_flag.update", ResourceKey: "osd-flag", Parameters: parameters})
		if (err == nil) != tc.valid {
			t.Fatalf("%+v err=%v", tc, err)
		}
	}
}
