package ceph

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"reflect"
	"testing"
)

func TestCollectProgressHistory(t *testing.T) {
	for _, body := range []string{`{"events":[],"completed":[]}`, `{"events":[{"id":"a","progress":0.2,"refs":{"origin":"rbd_support"}}],"completed":[{"id":"b","started_at":1,"finished_at":2,"failed":true,"failure_message":"failed"}]}`, `{}`, `null`, `{"events":[null],"completed":[]}`} {
		var calls []executor.CommandSpec
		provider := NativeProvider{Executor: recordingExecutor{base: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.overview_progress": []byte(body)}}, calls: &calls}}
		rows, err := provider.Collect(context.Background(), ClusterAccess{}, "fast")
		if err != nil {
			t.Fatal(err)
		}
		var history *cephdomain.ProgressHistory
		for _, row := range rows {
			if row.Kind == "overview" {
				history = row.Payload.(cephdomain.Overview).ProgressHistory
			}
		}
		valid := body == `{"events":[],"completed":[]}` || len(body) > 100
		if (history != nil) != valid {
			t.Fatalf("unexpected history for %s: %#v", body, history)
		}
		if history != nil && len(history.Completed) > 0 && (history.Completed[0]["failed"] != true || history.Events[0]["refs"].(map[string]any)["origin"] != "rbd_support") {
			t.Fatalf("lost fields %#v", history)
		}
		found := false
		for _, call := range calls {
			if call.ID == "collect.overview_progress" {
				found = true
				if !reflect.DeepEqual(call.Args, []string{"progress", "json"}) {
					t.Fatalf("args %v", call.Args)
				}
			}
		}
		if !found {
			t.Fatal("progress command missing")
		}
	}
}
