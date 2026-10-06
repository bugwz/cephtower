package mutation

import (
	"encoding/json"
	"testing"
)

func TestOSDActionTargetValidation(t *testing.T) {
	for _, action := range []string{"in", "out", "down", "reweight", "scrub", "deep-scrub"} {
		for _, id := range []string{"all", "any", "*", "--help", "-1", "01", "2147483648", "", "0", "2147483647"} {
			_, err := build(Request{Action: "osd.action", ResourceKey: "osd/" + id + "/action"}, map[string]any{"action": action, "weight": json.Number("0.5")})
			valid := id == "0" || id == "2147483647"
			if (err == nil) != valid {
				t.Fatalf("action=%s id=%q err=%v", action, id, err)
			}
		}
	}
}

func TestOSDReweightRange(t *testing.T) {
	for _, tc := range []struct {
		value string
		valid bool
	}{
		{"0", true}, {"1", true}, {"0.25", true}, {"1e-3", true},
		{"-0.1", false}, {"1.01", false}, {"NaN", false}, {"Inf", false},
		{"-Inf", false}, {"1e999", false}, {"ssd", false}, {"", false},
	} {
		_, err := build(Request{Action: "osd.action", ResourceKey: "osd/0/action"}, map[string]any{"action": "reweight", "weight": json.Number(tc.value)})
		if (err == nil) != tc.valid {
			t.Fatalf("weight=%q err=%v", tc.value, err)
		}
	}
}
