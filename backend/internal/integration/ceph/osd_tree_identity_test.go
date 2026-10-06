package ceph

import (
	"context"
	"testing"
)

func TestOSDTreeRejectsInvalidIdentities(t *testing.T) {
	for _, raw := range []string{
		`{}`, `null`, `{"nodes":null}`, `{"nodes":[null]}`,
		`{"nodes":[{"name":"osd.0","type":"osd"}]}`,
		`{"nodes":[{"id":null,"name":"osd.0","type":"osd"}]}`,
		`{"nodes":[{"id":0,"name":"osd.0"}]}`,
		`{"nodes":[{"id":0,"name":"osd.0","type":" osd"}]}`,
		`{"nodes":[{"id":0,"name":" ","type":"osd"}]}`,
		`{"nodes":[{"id":-1,"name":"osd.0","type":"osd"}]}`,
		`{"nodes":[{"id":0,"name":"osd.0","type":"osd"},{"id":0,"name":"osd.0","type":"osd"}]}`,
	} {
		t.Run(raw, func(t *testing.T) {
			p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.osd_tree": []byte(raw)}}}
			rows, err := p.Collect(context.Background(), ClusterAccess{}, "storage")
			if err == nil || rows != nil {
				t.Fatalf("malformed tree accepted: rows=%v err=%v", rows, err)
			}
		})
	}
}

func TestOSDTreeAcceptsEmptyAndZeroIdentities(t *testing.T) {
	for _, raw := range []string{`{"nodes":[]}`, `{"nodes":[{"id":-1,"name":"default","type":"root","children":[0]},{"id":0,"name":"osd.0","type":"osd"}]}`} {
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.osd_tree": []byte(raw)}}}
		if _, err := p.Collect(context.Background(), ClusterAccess{}, "storage"); err != nil {
			t.Fatal(err)
		}
	}
}
