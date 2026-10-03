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

func TestBucketACLIsReadOnlyInContract(t *testing.T) {
	found := false
	for _, parameter := range routeParameters(router.Route{Method: "GET", Path: "/rgw/bucket/policy"}) {
		if parameter.Name == "kind" {
			for _, value := range parameter.Enum {
				if value == "acl" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Fatal("ACL query missing")
	}
	for _, method := range []string{"PATCH", "DELETE"} {
		schema, _ := requestSchema(router.Route{Method: method, Path: "/rgw/bucket/policy"})
		for _, value := range schema.Fields["kind"].Enum {
			if value == "acl" {
				t.Fatal("read-only ACL writable")
			}
		}
	}
}

func TestBucketReplicationReadContract(t *testing.T) {
	found := false
	for _, parameter := range routeParameters(router.Route{Method: "GET", Path: "/rgw/bucket/policy"}) {
		if parameter.Name == "kind" {
			for _, value := range parameter.Enum {
				if value == "replication" {
					found = true
				}
			}
		}
	}
	if !found {
		t.Fatal("replication query missing")
	}
	for _, method := range []string{"PATCH", "DELETE"} {
		schema, _ := requestSchema(router.Route{Method: method, Path: "/rgw/bucket/policy"})
		allowed := false
		for _, value := range schema.Fields["kind"].Enum {
			if value == "replication" {
				allowed = true
			}
		}
		if allowed != (method == "DELETE") {
			t.Fatal("replication must allow deletion but not XML writes")
		}
	}
}
