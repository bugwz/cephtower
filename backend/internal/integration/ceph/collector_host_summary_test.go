package ceph

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/json"
	"testing"
)

func TestHostNativeSummaries(t *testing.T) {
	for _, raw := range []string{
		`[{"hostname":"node1","server":"Vendor Model","cpu_summary":"64C/128T","ram":"256 GiB","hdd_summary":"12/120 TB","ssd_summary":"-","os":"Linux"}]`,
		`[{"hostname":"node1"}]`,
		`[{"hostname":"node1","nic_count":0}]`,
		`[{"hostname":"node1","nic_count":4}]`,
		`[{"hostname":"node1","nic_count":"N/A"}]`,
	} {
		provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.host": []byte(raw)}}}
		rows, err := provider.Collect(context.Background(), ClusterAccess{}, "topology")
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, row := range rows {
			if row.Kind != "host" {
				continue
			}
			found = true
			payload := row.Payload.(cephdomain.Host)
			var input []map[string]any
			if err := json.Unmarshal([]byte(raw), &input); err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(payload.NativeSummary)
			if err != nil {
				t.Fatal(err)
			}
			var summary map[string]any
			if err := json.Unmarshal(encoded, &summary); err != nil {
				t.Fatal(err)
			}
			for _, key := range []string{"server", "cpu_summary", "ram", "hdd_summary", "ssd_summary", "os", "nic_count"} {
				if summary[key] != input[0][key] {
					t.Fatalf("%s: %v != %v", key, summary[key], input[0][key])
				}
			}
		}
		if !found {
			t.Fatal("host missing")
		}
	}
}
