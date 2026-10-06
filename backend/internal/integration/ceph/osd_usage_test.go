package ceph

import (
	"context"
	"testing"
)

func TestOSDUsageNativeValues(t *testing.T) {
	p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.osd_usage": []byte(`{"nodes":[{"id":-1,"type":"root","kb":100},{"id":0,"type":"osd","kb":18446744073709551615,"kb_used":0,"kb_avail":9007199254740993,"kb_used_data":1,"kb_used_omap":2,"kb_used_meta":3,"pgs":4,"utilization":0},{"id":1,"type":"osd"}]}`)}}}
	got := p.collectOSDUsage(context.Background(), ClusterAccess{})
	if len(got) != 2 || got[0] == nil || got[1] == nil {
		t.Fatalf("usage=%+v", got)
	}
	row := got[0]
	if *row.KB != "18446744073709551615" || *row.KBUsed != "0" || *row.KBAvailable != "9007199254740993" || *row.KBData != "1" || *row.KBOmap != "2" || *row.KBMetadata != "3" || *row.PGs != "4" || *row.Utilization != 0 {
		t.Fatalf("usage=%+v", row)
	}
	if got[1].KB != nil || got[1].PGs != nil {
		t.Fatal("invented values")
	}
}

func TestOSDUsageRejectsMalformedValues(t *testing.T) {
	for _, raw := range []string{`{}`, `{"nodes":null}`, `{"nodes":[{"type":"osd"}]}`, `{"nodes":[{"id":0,"type":"osd"},{"id":0,"type":"osd"}]}`, `{"nodes":[{"id":0,"type":"osd","kb":-1}]}`, `{"nodes":[{"id":0,"type":"osd","utilization":101}]}`} {
		p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.osd_usage": []byte(raw)}}}
		if got := p.collectOSDUsage(context.Background(), ClusterAccess{}); got != nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}
