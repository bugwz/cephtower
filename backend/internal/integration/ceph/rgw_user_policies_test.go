package ceph

import (
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"reflect"
	"testing"
	"time"
)

func TestRGWUserManagedPolicyCommandScope(t *testing.T) {
	for _, tc := range []struct {
		details string
		want    bool
	}{
		{`{"full_user_id":"tenant$user","account_id":"RGW123","type":"rgw"}`, true},
		{`{"full_user_id":"tenant$user","account_id":"RGW123","type":"root"}`, false},
		{`{"full_user_id":"tenant$user","account_id":"","type":"rgw"}`, false},
	} {
		var calls []executor.CommandSpec
		p := NativeProvider{Executor: recordingExecutor{calls: &calls, base: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
			"collect.rgw_user":          []byte(`["tenant$user"]`),
			"collect.rgw_user_detail":   []byte(tc.details),
			"collect.rgw_user_policies": []byte(`[]`),
		}}}}
		p.collectRGWOptional(context.Background(), ClusterAccess{}, time.Now())
		count := 0
		for _, call := range calls {
			if call.ID != "collect.rgw_user_policies" {
				continue
			}
			count++
			if call.Binary != executor.BinaryRGWAdmin || !reflect.DeepEqual(call.Args, []string{"user", "policy", "list", "attached", "--uid", "tenant$user", "--format", "json"}) {
				t.Fatalf("command = %+v", call)
			}
		}
		if (count == 1) != tc.want || count > 1 {
			t.Fatalf("policy calls = %d, wanted = %v", count, tc.want)
		}
	}
}

func TestRGWUserManagedPolicies(t *testing.T) {
	for _, tc := range []struct {
		name, response string
		valid          bool
	}{
		{"populated", `["arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"]`, true},
		{"empty", `[]`, true},
		{"duplicate", `["arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess","arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"]`, false},
		{"null", `null`, false},
		{"object", `{"AttachedPolicies":[]}`, false},
		{"empty ARN", `[""]`, false},
		{"null ARN", `[null]`, false},
		{"number", `[1]`, false},
		{"failure", `invalid`, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			trace := &collectionTrace{unavailable: map[string]struct{}{}}
			ctx := context.WithValue(context.Background(), collectionTraceKey{}, trace)
			p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
				"collect.rgw_user":          []byte(`["tenant$user"]`),
				"collect.rgw_user_detail":   []byte(`{"full_user_id":"tenant$user","account_id":"RGW123","type":"rgw"}`),
				"collect.rgw_user_policies": []byte(tc.response),
			}}}
			found := false
			for _, row := range p.collectRGWOptional(ctx, ClusterAccess{}, time.Now()) {
				if row.Kind != "rgw_user" {
					continue
				}
				found = true
				value, exists := row.Payload.(map[string]any)["managed_user_policies"]
				if exists != tc.valid {
					t.Fatalf("published policies = %v, valid = %v", value, tc.valid)
				}
				if tc.name == "empty" && !reflect.DeepEqual(value, []string{}) {
					t.Fatalf("empty policies = %#v", value)
				}
				if tc.name == "populated" && !reflect.DeepEqual(value, []string{"arn:aws:iam::aws:policy/AmazonS3ReadOnlyAccess"}) {
					t.Fatalf("policies = %#v", value)
				}
			}
			if !found {
				t.Fatal("user missing")
			}
			_, unavailable := trace.unavailable["rgw_user"]
			if !tc.valid && !unavailable {
				t.Fatalf("unavailable = %v, valid = %v", unavailable, tc.valid)
			}
		})
	}
}
