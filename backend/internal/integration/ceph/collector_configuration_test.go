package ceph

import (
	"context"
	"encoding/json"
	"testing"
)

func TestConfigurationRejectsUnknownOrAmbiguousValues(t *testing.T) {
	for _, response := range []string{
		`null`, `{}`, `[null]`,
		`[{"section":"global","name":"test"}]`,
		`[{"section":"global","name":"test","value":null}]`,
		`[{"section":"global","name":"test","value":0}]`,
		`[{"section":"global","name":"test","value":"a"},{"section":"global","name":"test","value":"b"}]`,
		`[{"section":"osd","mask":"class:ssd","name":"test","value":"a"},{"section":"osd","location_type":"class","location_value":"ssd","name":"test","value":"a"}]`,
	} {
		t.Run(response, func(t *testing.T) {
			provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.config": []byte(response)}}}
			rows, err := provider.Collect(context.Background(), ClusterAccess{}, "configuration")
			if err == nil || len(rows) != 0 {
				t.Fatalf("invalid configuration returned rows=%d err=%v", len(rows), err)
			}
		})
	}
}

func TestConfigurationPreservesExplicitEmptyAndScopedValues(t *testing.T) {
	for _, response := range []string{`[]`, `[
		{"section":"global","name":"test","value":""},
		{"section":"osd","name":"test","value":"0"},
		{"section":"osd","mask":"class:ssd","name":"test","value":"false"}
	]`} {
		provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.config": []byte(response)}}}
		rows, err := provider.Collect(context.Background(), ClusterAccess{}, "configuration")
		if err != nil {
			t.Fatal(err)
		}
		got := map[string]string{}
		for _, row := range rows {
			if row.Kind != "config_value" {
				continue
			}
			encoded, err := json.Marshal(row.Payload)
			if err != nil {
				t.Fatal(err)
			}
			var payload map[string]any
			if err := json.Unmarshal(encoded, &payload); err != nil {
				t.Fatal(err)
			}
			value, ok := payload["value"].(string)
			if !ok {
				t.Fatalf("missing explicit value: %s", encoded)
			}
			got[payload["who"].(string)] = value
		}
		if response == `[]` {
			if len(got) != 0 {
				t.Fatal(got)
			}
		} else if len(got) != 3 || got["global"] != "" || got["osd"] != "0" || got["osd/class:ssd"] != "false" {
			t.Fatal(got)
		}
	}
}
