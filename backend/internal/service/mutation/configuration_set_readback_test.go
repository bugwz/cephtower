package mutation

import (
	"cephtower/backend/internal/integration/ceph/executor"
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
		wantCommands := 2
		if !tc.success {
			wantCommands = 3
		}
		if len(runner.specs) != wantCommands || !runner.specs[0].Mutating || runner.specs[1].Mutating || !tc.success && runner.specs[2].Mutating {
			t.Fatalf("specs=%+v", runner.specs)
		}
	}
}

func TestConfigurationNativeNumericEquivalence(t *testing.T) {
	for _, tc := range []struct {
		kind, input, output string
		equal               bool
	}{
		{"size", "4G", "4294967296", true}, {"size", "4GiB", "4294967296", true},
		{"float", "0.5", "0.500000", true}, {"float", "0", "0.000000", true},
		{"float", "0.5", "0.5000001", false}, {"float", "NaN", "nan", false},
		{"int", "1K", "1000", true}, {"uint", "001", "1", true},
		{"bool", "TRUE", "true", true}, {"bool", "0", "false", true}, {"bool", "-2", "true", true},
		{"str", "001", "1", false}, {"str", "TRUE", "true", false},
		{"size", "4G", "4000000000", false}, {"uint", "-1", "18446744073709551615", false},
		{"int", "9223372036854775808", "9223372036854775809", false},
		{"bool", "yes", "true", false}, {"bool", "2147483648", "true", false},
	} {
		if got := configurationEquivalent(tc.input, tc.output, tc.kind); got != tc.equal {
			t.Fatalf("%+v got=%v", tc, got)
		}
	}
}

type normalizedConfigurationExecutor struct {
	specs    []executor.CommandSpec
	metadata string
}

func (e *normalizedConfigurationExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	if spec.ID == "config_value.set.metadata" {
		return executor.CommandResult{Stdout: []byte(e.metadata)}, nil
	}
	return executor.CommandResult{Stdout: []byte(`[{"section":"osd","name":"osd_memory_target","value":"4294967296"}]`)}, nil
}

func TestConfigurationNormalizedReadbackRequiresNativeType(t *testing.T) {
	service, _, id := newCephUserService(t)
	for _, metadata := range []string{`{"name":"osd_memory_target","type":"size"}`, `{"name":"osd_memory_target","type":"str"}`, `{"name":"foreign","type":"size"}`, `{}`} {
		runner := &normalizedConfigurationExecutor{metadata: metadata}
		service.executor = runner
		_, err := service.Execute(context.Background(), Request{ClusterID: id, Action: "config_value.set", ResourceKey: configurationTestKey("osd", "osd_memory_target"), Parameters: map[string]any{"value": "4G"}})
		if (err == nil) != (metadata == `{"name":"osd_memory_target","type":"size"}`) {
			t.Fatalf("metadata=%s err=%v", metadata, err)
		}
		if len(runner.specs) != 3 || runner.specs[2].Mutating {
			t.Fatalf("specs=%+v", runner.specs)
		}
	}
}
