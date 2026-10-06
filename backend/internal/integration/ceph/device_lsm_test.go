package ceph

import (
	cephdomain "cephtower/backend/internal/domain/ceph"
	"context"
	"encoding/json"
	"testing"
)

func TestCollectDeviceLSMData(t *testing.T) {
	provider := NativeProvider{Executor: malformedExecutor{base: fixtureExecutor{t}, override: map[string][]byte{"collect.device": []byte(`[{"hostname":"node1","path":"/dev/sda","lsm_data":{"serialNum":"abc","transport":"SAS","mediaType":"HDD","health":"Good","rpm":7200,"linkSpeed":12000,"unexpected":"not retained","ledSupport":{"IDENTsupport":"Supported","IDENTstatus":"Off","FAILsupport":"Unsupported","FAILstatus":"Unknown"}}}]`)}}}
	rows, err := provider.Collect(context.Background(), ClusterAccess{}, "inventory")
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatal(rows)
	}
	data := rows[0].Payload.(cephdomain.Device).LSMData
	if data["health"] != "Good" || data["rpm"] != "7200" || data["linkSpeed"] != "12000" {
		t.Fatal(data)
	}
	if _, ok := data["unexpected"]; ok {
		t.Fatal(data)
	}
	encoded, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["ledSupport"].(map[string]any)["FAILstatus"] != "Unknown" {
		t.Fatal(decoded)
	}
	if deviceLSMData(nil) != nil || len(deviceLSMData(map[string]any{"health": true})) != 0 {
		t.Fatal("invalid LSM data accepted")
	}
}
