package main

import (
	"cephtower/backend/internal/api/v1/handler"
	"cephtower/backend/internal/api/v1/router"
	"reflect"
	"testing"
)

func TestSilenceUpdateContract(t *testing.T) {
	got, ok := requestSchema(router.Route{Method: "PATCH", Path: "/alert/silence"})
	want, exists := handler.MutationRequestContract("silence.update")
	if !ok || !exists || !reflect.DeepEqual(got, want) {
		t.Fatal("silence update schema mismatch")
	}
	for _, field := range []string{"cluster_id", "silence_id", "expected_updated_at", "matchers", "startsAt", "endsAt", "createdBy", "comment"} {
		if !got.Fields[field].Required {
			t.Fatalf("%s not required", field)
		}
	}
	if err := handler.ValidateMutationRequest("silence.update", map[string]any{"cluster_id": float64(1)}); err == nil {
		t.Fatal("accepted incomplete update")
	}
}
