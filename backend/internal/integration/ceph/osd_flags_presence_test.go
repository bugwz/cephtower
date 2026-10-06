package ceph

import (
	"context"
	"reflect"
	"testing"
)

func TestOSDFlagsPresence(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want []string
	}{
		{`{"osds":[]}`, nil},
		{`{"osds":[],"flags":null}`, nil},
		{`{"osds":[],"flags":""}`, []string{}},
		{`{"osds":[],"flags":"noout,noscrub"}`, []string{"noout", "noscrub"}},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.osd_dump": []byte(tc.raw)}}}
			rows, err := p.Collect(context.Background(), ClusterAccess{}, "storage")
			if err != nil {
				t.Fatal(err)
			}
			for _, row := range rows {
				if row.Kind == "osd_flag" {
					if got := row.Payload.(map[string]any)["flags"]; !reflect.DeepEqual(got, tc.want) {
						t.Fatalf("flags=%#v want %#v", got, tc.want)
					}
					return
				}
			}
			t.Fatal("missing flags observation")
		})
	}
}
