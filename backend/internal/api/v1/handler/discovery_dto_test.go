package handler

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"cephtower/backend/internal/store"
)

func TestResourceDTOParsesInternalDataWithoutExposingStorageFields(t *testing.T) {
	now := time.Now().UTC()
	configured := `{"address":"configured-address","owner":"operator"}`
	row := store.CephEntityRecord{
		Kind: "host", NaturalKey: "ceph-node-1", ResourceVersion: 1, Source: "ceph_cli",
		ObservedAt: now, CreatedAt: now, UpdatedAt: now,
		ConfiguredData: &configured, DiscoveredData: `{"address":"discovered-address","device_class":"ssd"}`,
	}
	dto := toResourceDTO(row)
	data, ok := dto.Data.(map[string]any)
	if !ok || data["address"] != "configured-address" || data["device_class"] != "ssd" || data["owner"] != "operator" {
		t.Fatalf("resource data = %#v", dto.Data)
	}
	assertInternalDiscoveryFieldsHidden(t, dto)
	assertInternalDiscoveryFieldsHidden(t, row)
}

func TestClusterDTOParsesDiscoveryWithoutExposingStorageField(t *testing.T) {
	now := time.Now().UTC()
	row := store.CephCluster{
		ID: 1, Name: "fixture", MonitorAddresses: "mon:6789", ClientUsername: "client.fixture",
		DiscoveredData: `{"fsid":"00000000-0000-0000-0000-000000000001","version":"ceph version 20.2.2","status":"unavailable","error_code":"probe_failed","error_message":"unreachable"}`,
		Status:         "unknown", CreatedAt: now, UpdatedAt: now,
	}
	dto := toClusterDTO(row)
	if dto.FSID != "00000000-0000-0000-0000-000000000001" || dto.CephVersion != "20.2.2" || dto.Status != "unavailable" || dto.LastErrorCode != "probe_failed" {
		t.Fatalf("cluster DTO = %#v", dto)
	}
	assertInternalDiscoveryFieldsHidden(t, dto)
	assertInternalDiscoveryFieldsHidden(t, row)
}

func assertInternalDiscoveryFieldsHidden(t *testing.T, value any) {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"discovered_data", "configured_data"} {
		if strings.Contains(string(encoded), `"`+field+`"`) {
			t.Fatalf("internal field %q leaked in %s", field, encoded)
		}
	}
}
