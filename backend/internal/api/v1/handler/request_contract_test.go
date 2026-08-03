package handler

import "testing"

func TestMutationContractsRejectUnknownAndWrongType(t *testing.T) {
	if err := ValidateMutationRequest("host.create", map[string]any{"hostname": "node-1", "password": "secret"}); err == nil {
		t.Fatal("unknown field was accepted")
	}
	if err := ValidateMutationRequest("manager_module.update", map[string]any{"enabled": "true"}); err == nil {
		t.Fatal("wrong field type was accepted")
	}
	if err := ValidateMutationRequest("service.create", map[string]any{"service_type": "mgr", "placement": map[string]any{"shell": "bad"}}); err == nil {
		t.Fatal("unknown nested field was accepted")
	}
	if err := ValidateMutationRequest("service.create", map[string]any{"cluster_id": float64(1), "service_type": "mgr", "placement": map[string]any{"count": float64(2)}}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMutationRequest("device.zap", map[string]any{"cluster_id": float64(1), "host": "node-1", "device": "/dev/sdb"}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMutationRequest("device.zap", map[string]any{"cluster_id": float64(1), "device_id": "encoded"}); err == nil {
		t.Fatal("device zap without host and device was accepted")
	}
	if err := ValidateMutationRequest("pool.create", map[string]any{
		"cluster_id":        float64(1),
		"name":              "data",
		"pool_type":         "replicated",
		"pg_num":            float64(32),
		"pg_autoscale_mode": "on",
		"size":              float64(3),
		"applications":      []any{"rbd"},
		"crush_rule":        "replicated_rule",
		"compression_mode":  "none",
		"quota_max_bytes":   float64(0),
		"quota_unit":        "GiB",
		"quota_max_objects": float64(0),
	}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateMutationRequest("pool.update", map[string]any{
		"cluster_id": float64(1),
		"pool":       "data",
		"field":      "crush_rule",
		"value":      "replicated_rule",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestMutationContractsCoverAllRegisteredActions(t *testing.T) {
	for _, action := range MutationContractActions() {
		contract, ok := MutationRequestContract(action)
		if !ok || contract.Fields == nil {
			t.Fatalf("action %s has no contract", action)
		}
	}
}
