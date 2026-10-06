package mutation

import (
	"context"
	"testing"
)

func TestConfigurationSetReadback(t *testing.T) {
	key := configurationTestKey("osd", "osd_max_backfills")
	for _, tc := range []struct {
		raw, value string
		matches    bool
	}{
		{`[{"section":"osd","name":"osd_max_backfills","value":"1"}]`, "1", true},
		{`[{"section":"osd","name":"osd_max_backfills","value":"0"}]`, "0", true},
		{`[{"section":"osd","name":"osd_max_backfills","value":""}]`, "", true},
		{`[{"section":"osd","name":"osd_max_backfills","value":"2"}]`, "1", false},
		{`[{"section":"global","name":"osd_max_backfills","value":"1"}]`, "1", false},
		{`[{"section":"osd","name":"osd_max_backfills","value":"1","mask":"host:node1"}]`, "1", false},
		{`[{"section":"osd","name":"osd_max_backfills","value":null}]`, "1", false},
		{`[{"section":"osd","name":"osd_max_backfills","value":"1"},{"section":"osd","name":"osd_max_backfills","value":"1"}]`, "1", false},
		{`[]`, "1", false}, {`null`, "1", false}, {`[] {}`, "1", false},
	} {
		if got := configurationSet(key, []byte(tc.raw), map[string]any{"value": tc.value}); got != tc.matches {
			t.Fatalf("raw=%s got=%v", tc.raw, got)
		}
	}
	masked := configurationTestKey("osd/host:node1/class:ssd", "osd_max_backfills")
	if !configurationSet(masked, []byte(`[{"section":"osd","name":"osd_max_backfills","value":"4","location_type":"host","location_value":"node1","device_class":"ssd"}]`), map[string]any{"value": "4"}) {
		t.Fatal("masked target not matched")
	}
}

func TestConfigurationSetRequiresVerifiedReadback(t *testing.T) {
	service, _, id := newCephUserService(t)
	runner := &osdSafetyExecutor{}
	service.executor = runner
	request := Request{ClusterID: id, Action: "config_value.set", ResourceKey: configurationTestKey("osd", "osd_max_backfills"), Parameters: map[string]any{"value": "1"}}
	for _, tc := range []struct {
		output  string
		success bool
	}{
		{`[{"section":"osd","name":"osd_max_backfills","value":"1"}]`, true},
		{`[{"section":"osd","name":"osd_max_backfills","value":"2"}]`, false},
		{`[]`, false},
	} {
		runner.output = tc.output
		runner.specs = nil
		_, err := service.Execute(context.Background(), request)
		if (err == nil) != tc.success {
			t.Fatalf("output=%s err=%v", tc.output, err)
		}
		if len(runner.specs) != 2 || !runner.specs[0].Mutating || runner.specs[1].Mutating {
			t.Fatalf("specs=%+v", runner.specs)
		}
	}
}
