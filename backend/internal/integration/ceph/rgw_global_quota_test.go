package ceph

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

func TestGlobalRGWQuotaCollection(t *testing.T) {
	for _, tc := range []struct {
		raw   string
		valid bool
	}{
		{`{"user quota":{"enabled":true,"max_size":0,"max_objects":-1},"bucket quota":{"enabled":false,"max_size":2048,"max_objects":10}}`, true},
		{`{"user_quota":{},"bucket_quota":{}}`, false},
		{`{"user quota":{},"bucket quota":null}`, false},
		{`{"user quota":[],"bucket quota":{}}`, false},
		{`{}`, false}, {`null`, false}, {`invalid`, false},
	} {
		var calls []executor.CommandSpec
		provider := NativeProvider{Executor: recordingExecutor{base: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.rgw_global_quota": []byte(tc.raw)}}, calls: &calls}}
		rows, err := provider.collectStorage(context.Background(), ClusterAccess{})
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, row := range rows {
			if row.Kind != "rgw_status" {
				continue
			}
			found = true
			payload := row.Payload.(cephdomain.RGWStatus)
			if (payload.GlobalQuota != nil) != tc.valid {
				t.Fatalf("payload=%#v for %s", payload, tc.raw)
			}
			if tc.valid {
				raw, _ := json.Marshal(payload)
				var api map[string]json.RawMessage
				if json.Unmarshal(raw, &api) != nil || len(api["global_quota"]) == 0 {
					t.Fatalf("missing API field: %s", raw)
				}
				user := payload.GlobalQuota["user_quota"].(map[string]any)
				if user["enabled"] != true || user["max_size"] != json.Number("0") || user["max_objects"] != json.Number("-1") {
					t.Fatalf("lost quota semantics: %#v", user)
				}
			}
		}
		if !found {
			t.Fatal("missing RGW status")
		}
		count := 0
		for _, call := range calls {
			if call.ID != "collect.rgw_global_quota" {
				continue
			}
			count++
			if call.Binary != executor.BinaryRGWAdmin || call.Mutating || !reflect.DeepEqual(call.Args, []string{"global", "quota", "get", "--format", "json"}) {
				t.Fatalf("wrong command: %#v", call)
			}
		}
		if count != 1 {
			t.Fatalf("command count=%d", count)
		}
	}
	if !reflect.DeepEqual(collectionFailureKinds["collect.rgw_global_quota"], []string{"rgw_status"}) {
		t.Fatal("missing failure tracking")
	}
}
