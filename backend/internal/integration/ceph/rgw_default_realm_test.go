package ceph

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
)

func TestRGWStatusDefaultRealmIdentity(t *testing.T) {
	for _, tc := range []struct {
		field string
		want  *string
	}{
		{`,"default_info":"realm-id"`, strPtr("realm-id")},
		{`,"default_info":""`, strPtr("")},
		{"", nil}, {`,"default_info":null`, nil}, {`,"default_info":7`, nil},
		{`,"default_info":false`, nil}, {`,"default_info":{}`, nil},
	} {
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.rgw_status": []byte(`{"realms":["east","west"]` + tc.field + `}`)}}}
		rows, err := p.collectStorage(context.Background(), ClusterAccess{})
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
			if !reflect.DeepEqual(payload.DefaultRealmID, tc.want) || !reflect.DeepEqual(payload.Realms, []string{"east", "west"}) {
				t.Fatalf("payload=%#v", payload)
			}
			raw, _ := json.Marshal(payload)
			var wire map[string]json.RawMessage
			if json.Unmarshal(raw, &wire) != nil {
				t.Fatal("invalid payload")
			}
			want, _ := json.Marshal(tc.want)
			if string(wire["default_realm_id"]) != string(want) {
				t.Fatalf("lost empty/unknown distinction: %s", raw)
			}
		}
		if !found {
			t.Fatal("missing status")
		}
	}
}

func strPtr(value string) *string { return &value }
