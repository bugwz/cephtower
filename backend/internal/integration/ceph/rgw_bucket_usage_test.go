package ceph

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestRGWBucketUsagePrecision(t *testing.T) {
	p := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{
		"collect.rgw_bucket":        []byte(`["team/photos"]`),
		"collect.rgw_bucket_detail": []byte(`{"bucket":"photos","tenant":"team","usage":{"rgw.main":{"size":18446744073709551615,"size_actual":9007199254740993,"size_utilized":0,"num_objects":9007199254740995},"rgw.multimeta":{"num_objects":2}}}`),
	}}}
	for _, row := range p.collectRGWOptional(context.Background(), ClusterAccess{}, time.Now()) {
		if row.Kind != "rgw_bucket" {
			continue
		}
		encoded, err := json.Marshal(row.Payload)
		if err != nil {
			t.Fatal(err)
		}
		for _, exact := range []string{`"size":"18446744073709551615"`, `"size_actual":"9007199254740993"`, `"size_utilized":"0"`, `"num_objects":"9007199254740995"`, `"num_objects":"2"`} {
			if !strings.Contains(string(encoded), exact) {
				t.Fatal("counter lost precision", exact)
			}
		}
		return
	}
	t.Fatal("bucket missing")
}

func TestRGWBucketUsageDoesNotInventCounters(t *testing.T) {
	for _, raw := range []string{`{}`, `{"usage":null}`, `{"usage":{"rgw.main":{}}}`, `{"usage":{"rgw.main":{"num_objects":null,"unknown":5}}}`} {
		var data map[string]any
		decoder := json.NewDecoder(strings.NewReader(raw))
		decoder.UseNumber()
		if err := decoder.Decode(&data); err != nil {
			t.Fatal(err)
		}
		before, _ := json.Marshal(data)
		preserveRGWBucketUsage(data)
		after, _ := json.Marshal(data)
		if string(before) != string(after) {
			t.Fatal("missing data was manufactured")
		}
	}
}
