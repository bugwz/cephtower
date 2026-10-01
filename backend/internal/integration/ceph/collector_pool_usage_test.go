package ceph

import (
	"cephtower/backend/internal/integration/ceph/executor"
	"context"
	"reflect"
	"testing"
)

func TestCollectPoolUsage(t *testing.T) {
	var calls []executor.CommandSpec
	p := NativeProvider{Executor: recordingExecutor{base: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.pool_usage": []byte(`{"pools":[{"id":7,"stats":{"percent_used":0.25,"stored":1024,"bytes_used":3072,"max_avail":8192}},{"id":8,"stats":{"percent_used":0}}]}`),
	}}, calls: &calls}}
	got := p.collectPoolUsage(context.Background(), ClusterAccess{})
	if len(got) != 2 || *got[7].PercentUsed != 25 || *got[8].PercentUsed != 0 || *got[7].Stored != 1024 || *got[7].BytesUsed != 3072 || *got[7].MaxAvail != 8192 || got[8].Stored != nil {
		t.Fatalf("usage = %#v", got)
	}
	if len(calls) != 1 || !reflect.DeepEqual(calls[0].Args, []string{"df", "detail", "--format", "json"}) {
		t.Fatalf("commands = %#v", calls)
	}
	for _, response := range []string{`{}`, `{"pools":null}`, `{"pools":[{}]}`, `{"pools":[{"id":1},{"id":1}]}`, `{"pools":[{"id":1,"stats":{"percent_used":25}}]}`, `{"pools":[{"id":1,"stats":{"stored":-1}}]}`} {
		p.Executor = malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.pool_usage": []byte(response)}}
		if got := p.collectPoolUsage(context.Background(), ClusterAccess{}); got != nil {
			t.Fatalf("accepted invalid response %s", response)
		}
	}
}
