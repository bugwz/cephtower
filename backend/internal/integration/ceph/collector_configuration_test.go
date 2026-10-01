package ceph

import (
	"context"
	"encoding/json"
	"slices"
	"testing"
)

func TestConfigurationOptionsAreCollectedAsCompleteLists(t *testing.T) {
	for _, test := range []struct {
		response string
		valid    bool
		count    int
	}{
		{`[]`, true, 0},
		{`["osd_memory_target","mgr/dashboard/ssl_server_port"]`, true, 2},
		{`null`, false, 0}, {`{}`, false, 0}, {`[null]`, false, 0},
		{`["good",""]`, false, 0}, {`["good"," padded"]`, false, 0},
		{`["good","good"]`, false, 0}, {`["good",1]`, false, 0},
	} {
		provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
			"collect.config_option": []byte(test.response),
			"collect.config":        []byte(`[{"section":"global","name":"test","value":"0"}]`),
		}}}
		result, err := provider.CollectWithMetadata(context.Background(), ClusterAccess{}, "configuration")
		if err != nil {
			t.Fatal(err)
		}
		if slices.Contains(result.UnavailableKinds, "config_option") == test.valid {
			t.Fatalf("response=%s unavailable=%v", test.response, result.UnavailableKinds)
		}
		options, values := 0, 0
		for _, row := range result.Observations {
			if row.Kind == "config_option" {
				options++
			}
			if row.Kind == "config_value" {
				values++
			}
		}
		if options != test.count || values != 1 {
			t.Fatalf("response=%s options=%d values=%d", test.response, options, values)
		}
	}
}

func TestConfigurationRejectsUnknownOrAmbiguousValues(t *testing.T) {
	for _, response := range []string{
		`null`, `{}`, `[null]`,
		`[{"section":"global","name":"test"}]`,
		`[{"section":"global","name":"test","value":null}]`,
		`[{"section":"global","name":"test","value":0}]`,
		`[{"section":"osd","name":"test","value":"0","location_type":"host"}]`,
		`[{"section":"osd","name":"test","value":"0","location_type":"host","location_value":"a","mask":"host:b/class:ssd"}]`,
		`[{"section":"osd","name":"test","value":"0","device_class":"ssd","mask":"class:hdd"}]`,
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

func TestConfigurationCombinedRestrictionsReachInventory(t *testing.T) {
	for _, mask := range []string{``, `,"mask":"host:node1/class:ssd"`} {
		response := `[{"section":"osd","name":"test","value":"1","location_type":"host","location_value":"node1","device_class":"ssd"` + mask + `},{"section":"osd","name":"test","value":"2","location_type":"host","location_value":"node1","mask":"host:node1/class:hdd"}]`
		provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.config": []byte(response)}}}
		rows, err := provider.Collect(context.Background(), ClusterAccess{}, "configuration")
		if err != nil {
			t.Fatal(err)
		}
		got := map[string]map[string]any{}
		for _, row := range rows {
			if row.Kind != "config_value" {
				continue
			}
			data, err := json.Marshal(row.Payload)
			if err != nil {
				t.Fatal(err)
			}
			var payload map[string]any
			if err := json.Unmarshal(data, &payload); err != nil {
				t.Fatal(err)
			}
			got[row.NaturalKey] = payload
		}
		ssd := got["osd/host:node1/class:ssd:test"]
		hdd := got["osd/host:node1/class:hdd:test"]
		if len(got) != 2 || ssd["who"] != "osd/host:node1/class:ssd" || ssd["device_class"] != "ssd" || hdd["who"] != "osd/host:node1/class:hdd" {
			t.Fatal(got)
		}
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
