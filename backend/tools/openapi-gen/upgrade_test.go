package main

import (
	"reflect"
	"testing"

	"cephtower/backend/internal/api/v1/handler"
	"cephtower/backend/internal/api/v1/router"
)

func TestUpgradeSchemasUseRuntimeContracts(t *testing.T) {
	for path, action := range map[string]string{"/upgrade/check": "upgrade.check", "/upgrade/action": "upgrade.action"} {
		got, ok := requestSchema(router.Route{Method: "POST", Path: path})
		want, _ := handler.MutationRequestContract(action)
		if !ok || !reflect.DeepEqual(got, want) {
			t.Fatalf("%s schema differs from runtime contract", path)
		}
		if _, exists := got.Fields["password"]; exists {
			t.Fatal("unrelated mutation fields leaked into upgrade schema")
		}
		for _, field := range []string{"version", "image"} {
			if got.Fields[field].Type != "string" || got.Fields[field].Required {
				t.Fatalf("invalid target field %s", field)
			}
		}
	}
}
