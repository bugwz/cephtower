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
}

func TestMutationContractsCoverAllRegisteredActions(t *testing.T) {
	for _, action := range MutationContractActions() {
		contract, ok := MutationRequestContract(action)
		if !ok || contract.Fields == nil {
			t.Fatalf("action %s has no contract", action)
		}
	}
}
