package handler

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"cephtower/backend/internal/store"
)

func TestResourceDTOUsesOnlyDiscoveredData(t *testing.T) {
	now := time.Now().UTC()
	configured := `{"address":"configured-address","owner":"operator"}`
	row := store.CephEntityRecord{
		Kind: "host", NaturalKey: "ceph-node-1", ResourceVersion: 1, Source: "ceph_cli",
		ObservedAt: now, CreatedAt: now, UpdatedAt: now,
		ConfiguredData: &configured, DiscoveredData: `{"address":"discovered-address","device_class":"ssd"}`,
	}
	dto := toResourceDTO(row)
	data, ok := dto.Data.(map[string]any)
	if !ok || data["address"] != "discovered-address" || data["device_class"] != "ssd" || data["owner"] != nil {
		t.Fatalf("resource data = %#v", dto.Data)
	}
	assertInternalDiscoveryFieldsHidden(t, dto)
	assertInternalDiscoveryFieldsHidden(t, row)
}

func TestPoolResourceDTOPrefersDiscoveredClusterData(t *testing.T) {
	now := time.Now().UTC()
	configured := `{"quota_max_bytes":1024,"compression_mode":"passive","applications":["rbd"],"owner":"operator"}`
	row := store.CephEntityRecord{
		Kind: "pool", NaturalKey: "data", ResourceVersion: 1, Source: "ceph_cli",
		ObservedAt: now, CreatedAt: now, UpdatedAt: now,
		ConfiguredData: &configured, DiscoveredData: `{"quota_max_bytes":0,"compression_mode":"none"}`,
	}
	dto := toResourceDTO(row)
	data, ok := dto.Data.(map[string]any)
	if !ok || data["quota_max_bytes"] != float64(0) || data["compression_mode"] != "none" || data["owner"] != nil || data["applications"] != nil {
		t.Fatalf("pool resource data = %#v", dto.Data)
	}
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

func TestRBDStateComesOnlyFromDiscovery(t *testing.T) {
	configured := `{"name":"old-name","image_spec":"raw/pool/image","destination":"other/image"}`
	row := store.CephEntityRecord{Kind: "rbd_image", ConfiguredData: &configured, DiscoveredData: `{"name":"new-name","image_spec":"encoded"}`}
	data := toResourceDTO(row).Data.(map[string]any)
	if data["name"] != "new-name" || data["image_spec"] != "encoded" || data["destination"] != nil {
		t.Fatalf("request fields leaked into observed state: %v", data)
	}
}

func TestRGWUserUsesNativeState(t *testing.T) {
	configured := `{"suspended":false,"email":"old@example.com"}`
	row := store.CephEntityRecord{Kind: "rgw_user", ConfiguredData: &configured, DiscoveredData: `{"uid":"tenant$user","suspended":1,"email":""}`}
	data := toResourceDTO(row).Data.(map[string]any)
	if data["suspended"] != float64(1) || data["email"] != "" {
		t.Fatalf("native user state overwritten: %v", data)
	}
}
