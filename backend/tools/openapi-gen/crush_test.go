package main

import (
	"reflect"
	"testing"

	"cephtower/backend/internal/api/v1/handler"
	"cephtower/backend/internal/api/v1/router"
)

func TestPlacementSchemasMatchRuntime(t *testing.T) {
	for _, tc := range []struct{ method, path, action string }{
		{"POST", "/crush/rule", "crush_rule.create"},
		{"PATCH", "/crush/rule", "crush_rule.update"},
		{"DELETE", "/crush/rule", "crush_rule.delete"},
		{"POST", "/erasure/code/profile", "erasure_code_profile.create"},
		{"DELETE", "/erasure/code/profile", "erasure_code_profile.delete"},
	} {
		got, ok := requestSchema(router.Route{Method: tc.method, Path: tc.path})
		want, _ := handler.MutationRequestContract(tc.action)
		if !ok || !reflect.DeepEqual(got, want) {
			t.Fatalf("%s differs from runtime", tc.action)
		}
	}
	contract, _ := requestSchema(router.Route{Method: "PATCH", Path: "/crush/rule"})
	for _, field := range []string{"name", "new_name"} {
		if !contract.Fields[field].Required || contract.Fields[field].Type != "string" {
			t.Fatalf("missing required %s", field)
		}
	}
	if err := handler.ValidateMutationRequest("crush_rule.update", map[string]any{"cluster_id": float64(1), "name": "old"}); err == nil {
		t.Fatal("accepted missing destination")
	}
}
