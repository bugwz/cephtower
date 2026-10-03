package main

import (
	"reflect"
	"testing"

	"cephtower/backend/internal/api/v1/handler"
	"cephtower/backend/internal/api/v1/router"
)

func TestPoolMirrorScheduleSchemaUsesRuntimeContract(t *testing.T) {
	got, ok := requestSchema(router.Route{Method: "POST", Path: "/rbd/mirroring/schedule"})
	want, _ := handler.MutationRequestContract("rbd_mirroring.schedule")
	if !ok || !reflect.DeepEqual(got, want) {
		t.Fatal("pool schedule schema differs from runtime contract")
	}
	for _, key := range []string{"cluster_id", "pool", "action"} {
		if !got.Fields[key].Required {
			t.Fatalf("missing required field %s", key)
		}
	}
	if _, exists := got.Fields["password"]; exists {
		t.Fatal("unrelated fields leaked into schedule schema")
	}
	for _, key := range []string{"interval", "start_time"} {
		if _, exists := mutationFieldUnion()[key]; exists {
			t.Fatalf("schedule field %s leaked into unrelated routes", key)
		}
	}
}

func TestNamespaceScheduleSchemaUsesRuntimeContract(t *testing.T) {
	got, ok := requestSchema(router.Route{Method: "POST", Path: "/rbd/namespace/schedule"})
	want, _ := handler.MutationRequestContract("rbd_namespace.schedule")
	if !ok || !reflect.DeepEqual(got, want) || !got.Fields["namespace"].Required {
		t.Fatal("namespace schedule schema differs from runtime contract")
	}
}
