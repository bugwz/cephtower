package ceph

import (
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"reflect"
	"testing"
	"time"
)

func TestRGWBucketNamesStrict(t *testing.T) {
	for _, value := range []any{nil, map[string]any{"buckets": []any{}}, []any{nil}, []any{1}, []any{""}, []any{"a", "a"}} {
		if _, ok := rgwBucketNames(value); ok {
			t.Fatal("invalid enumeration accepted")
		}
	}
	for _, value := range [][]any{{}, {"team/photos", "photos"}} {
		if names, ok := rgwBucketNames(value); !ok || len(names) != len(value) {
			t.Fatal("valid enumeration rejected")
		}
	}
}

type bucketEnumerationExecutor struct {
	base  executor.Executor
	calls []executor.CommandSpec
}

func (e *bucketEnumerationExecutor) Run(ctx context.Context, access executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	if spec.ID == "collect.rgw_bucket" {
		e.calls = append(e.calls, spec)
	}
	return e.base.Run(ctx, access, spec)
}

func TestRGWBucketUsesMetadataEnumeration(t *testing.T) {
	runner := &bucketEnumerationExecutor{base: fixtureExecutor{t}}
	p := NativeProvider{Executor: runner}
	p.collectRGWOptional(context.Background(), ClusterAccess{}, time.Now())
	if len(runner.calls) != 1 || runner.calls[0].Binary != executor.BinaryRGWAdmin || runner.calls[0].Mutating || !reflect.DeepEqual(runner.calls[0].Args, []string{"metadata", "list", "bucket", "--format", "json"}) {
		t.Fatal("unsafe enumeration command", runner.calls)
	}
}

func TestRGWBucketInvalidEnumerationMarksUnavailable(t *testing.T) {
	for _, raw := range []string{`null`, `{}`, `["photos",null]`, `["photos","photos"]`} {
		trace := &collectionTrace{unavailable: map[string]struct{}{}}
		ctx := context.WithValue(context.Background(), collectionTraceKey{}, trace)
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.rgw_bucket": []byte(raw)}}}
		rows := p.collectRGWOptional(ctx, ClusterAccess{}, time.Now())
		for _, row := range rows {
			if row.Kind == "rgw_bucket" {
				t.Fatal("partial bucket inventory emitted")
			}
		}
		if _, ok := trace.unavailable["rgw_bucket"]; !ok {
			t.Fatal("invalid list became available empty inventory")
		}
	}
}
