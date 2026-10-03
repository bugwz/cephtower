package main

import (
	"reflect"
	"testing"

	"cephtower/backend/internal/api/v1/handler"
	"cephtower/backend/internal/api/v1/router"
)

func TestBucketConfigurationDeletionSchemaMatchesRuntime(t *testing.T) {
	got, ok := requestSchema(router.Route{Method: "DELETE", Path: "/rgw/bucket/policy"})
	want, _ := handler.MutationRequestContract("rgw_bucket_policy.delete")
	if !ok || !reflect.DeepEqual(got, want) {
		t.Fatal("deletion contract differs")
	}
	for _, key := range []string{"cluster_id", "kind", "bucket_id"} {
		if !got.Fields[key].Required {
			t.Fatal("required identity missing: " + key)
		}
	}
	if _, ok := got.Fields["document"]; ok {
		t.Fatal("delete must not take a configuration document")
	}
}
