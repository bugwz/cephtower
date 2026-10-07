package ceph

import (
	"context"
	"testing"
	"time"
)

func TestMirrorStateCountsPreservePrecision(t *testing.T) {
	for _, tc := range []struct {
		raw  string
		want any
	}{
		{"0", "0"}, {"18446744073709551615", "18446744073709551615"},
		{"9007199254740993", "9007199254740993"}, {"-1", nil}, {"1.5", nil}, {"1e3", nil},
		{"18446744073709551616", nil}, {`"2"`, nil}, {"null", nil}, {"false", nil},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
				"collect.rbd_mirroring":        []byte(`{"mode":"image","peers":[]}`),
				"collect.rbd_mirroring_status": []byte(`{"summary":{"health":"OK","states":{"replaying":` + tc.raw + `}}}`),
			}}}
			rows := p.collectStorageOptional(context.Background(), ClusterAccess{}, []poolWire{{PoolName: "pool"}}, fsDumpWire{}, time.Now())
			for _, row := range rows {
				if row.Kind != "rbd_mirroring" {
					continue
				}
				summary := row.Payload.(map[string]any)["summary"].(map[string]any)
				if summary["states"].(map[string]any)["replaying"] != tc.want || summary["health"] != "OK" {
					t.Fatalf("summary=%#v", summary)
				}
				return
			}
			t.Fatal("missing pool summary")
		})
	}
}
