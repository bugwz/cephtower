package ceph

import (
	"context"
	"testing"
)

func TestOSDPerfSnapshots(t *testing.T) {
	p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.osd_perf": []byte(`{"osd_perf_infos":[{"id":0,"perf_stats":{"commit_latency_ms":0,"apply_latency_ms":0.125}},{"id":1,"perf_stats":{}}]}`)}}}
	got := p.collectOSDPerf(context.Background(), ClusterAccess{})
	if len(got) != 2 || *got[0].CommitLatencyMS != 0 || *got[0].ApplyLatencyMS != 0.125 || got[1].CommitLatencyMS != nil {
		t.Fatalf("perf=%+v", got)
	}
	for _, raw := range []string{`{}`, `{"osd_perf_infos":null}`, `{"osd_perf_infos":[null]}`, `{"osd_perf_infos":[{"id":0}]}`, `{"osd_perf_infos":[{"id":0,"perf_stats":{}},{"id":0,"perf_stats":{}}]}`, `{"osd_perf_infos":[{"id":0,"perf_stats":{"apply_latency_ms":-1}}]}`} {
		p.Executor = malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.osd_perf": []byte(raw)}}
		if p.collectOSDPerf(context.Background(), ClusterAccess{}) != nil {
			t.Fatalf("accepted %s", raw)
		}
	}
}
