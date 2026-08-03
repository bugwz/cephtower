package store

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"cephtower/backend/internal/config"
)

func TestListClustersAppliesMultiValueFiltersAndReturnsAllOptions(t *testing.T) {
	db, err := Open(config.DatabaseConfig{EncryptionKey: schemaTestKey, Engine: EngineSQLite, SQLite: config.SQLiteConfig{Name: "clusters.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer Close(db)
	ctx := context.Background()
	now := time.Now().UTC()
	clusters := []CephCluster{
		{Name: "alpha", MonitorAddresses: "alpha:6789", ClientUsername: "client.admin", ClientKey: "cipher", CreatedAt: now, UpdatedAt: now},
		{Name: "beta", MonitorAddresses: "beta:6789", ClientUsername: "client.readonly", ClientKey: "cipher", CreatedAt: now, UpdatedAt: now},
		{Name: "gamma", MonitorAddresses: "gamma:6789", ClientUsername: "client.admin", ClientKey: "cipher", CreatedAt: now, UpdatedAt: now},
	}
	for index := range clusters {
		if err := db.CreateCluster(ctx, &clusters[index]); err != nil {
			t.Fatal(err)
		}
	}

	filtered, err := db.ListClusters(ctx, ClusterFilter{
		Names:           []string{"alpha", "gamma"},
		ClientUsernames: []string{"client.admin"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := clusterNames(filtered); !reflect.DeepEqual(got, []string{"alpha", "gamma"}) {
		t.Fatalf("filtered cluster names = %v", got)
	}

	options, err := db.ClusterFilterOptions(ctx, []string{"name", "client_username"})
	if err != nil {
		t.Fatal(err)
	}
	expected := ClusterFilterOptions{
		"name":            {"alpha", "beta", "gamma"},
		"client_username": {"client.admin", "client.readonly"},
	}
	if !reflect.DeepEqual(options, expected) {
		t.Fatalf("cluster filter options = %#v, want %#v", options, expected)
	}
}

func TestListResourcesAppliesMultiValuePayloadFilters(t *testing.T) {
	db, err := Open(config.DatabaseConfig{EncryptionKey: schemaTestKey, Engine: EngineSQLite, SQLite: config.SQLiteConfig{Name: "resources.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer Close(db)
	ctx := context.Background()
	now := time.Now().UTC()
	cluster := CephCluster{Name: "c", MonitorAddresses: "mon:6789", ClientUsername: "client.test", ClientKey: "cipher", CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(ctx, &cluster); err != nil {
		t.Fatal(err)
	}
	active := "active"
	rows := []CephEntityRecord{
		resourceRecord(cluster.ID, "osd.a", "osd-a", &active, now, map[string]any{"device_class": "ssd", "labels": []string{"fast", "blue"}}),
		resourceRecord(cluster.ID, "osd.b", "osd-b", &active, now, map[string]any{"device_class": "hdd", "labels": []string{"bulk"}}),
		resourceRecord(cluster.ID, "osd.c", "osd-c", nil, now, map[string]any{"device_class": "nvme", "labels": []string{"fast"}}),
	}
	if err := db.ReconcileResources(ctx, cluster.ID, 1, rows, []string{"osd"}); err != nil {
		t.Fatal(err)
	}

	filtered, err := db.ListResources(ctx, cluster.ID, ResourceFilter{
		Kind: "osd",
		FieldValues: map[string][]string{
			"device_class": {"ssd", "nvme"},
			"labels":       {"fast"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := resourceKeys(filtered); !reflect.DeepEqual(got, []string{"osd.a", "osd.c"}) {
		t.Fatalf("filtered resource keys = %v", got)
	}
}

func TestResourceFilterOptionsComeFromFullResourceSet(t *testing.T) {
	db, err := Open(config.DatabaseConfig{EncryptionKey: schemaTestKey, Engine: EngineSQLite, SQLite: config.SQLiteConfig{Name: "filter-options.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer Close(db)
	ctx := context.Background()
	now := time.Now().UTC()
	cluster := CephCluster{Name: "c", MonitorAddresses: "mon:6789", ClientUsername: "client.test", ClientKey: "cipher", CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(ctx, &cluster); err != nil {
		t.Fatal(err)
	}
	active := "active"
	if err := db.ReconcileResources(ctx, cluster.ID, 1, []CephEntityRecord{
		resourceRecord(cluster.ID, "osd.a", "osd-a", &active, now, map[string]any{"device_class": "ssd", "labels": []string{"fast", "blue"}}),
		resourceRecord(cluster.ID, "osd.b", "osd-b", &active, now, map[string]any{"device_class": "hdd", "labels": []string{"bulk", "blue"}}),
	}, []string{"osd"}); err != nil {
		t.Fatal(err)
	}

	options, err := db.ResourceFilterOptions(ctx, cluster.ID, ResourceFilter{Kind: "osd"}, []string{"device_class", "labels", "status"})
	if err != nil {
		t.Fatal(err)
	}
	expected := ResourceFilterOptions{
		"device_class": {"hdd", "ssd"},
		"labels":       {"blue", "bulk", "fast"},
		"status":       {"active"},
	}
	if !reflect.DeepEqual(options, expected) {
		t.Fatalf("filter options = %#v, want %#v", options, expected)
	}
}

func TestHostReconciliationPreservesUserConfiguration(t *testing.T) {
	db, err := Open(config.DatabaseConfig{EncryptionKey: schemaTestKey, Engine: EngineSQLite, SQLite: config.SQLiteConfig{Name: "hosts.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer Close(db)
	ctx := context.Background()
	now := time.Now().UTC()
	cluster := CephCluster{Name: "c", MonitorAddresses: "mon:6789", ClientUsername: "client.test", ClientKey: "cipher", CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(ctx, &cluster); err != nil {
		t.Fatal(err)
	}
	host := CephHost{ClusterID: cluster.ID, Hostname: "ceph-node-1", SSHAddress: "192.0.2.10", SSHPort: 2222, SSHUser: "cephadmin", DiscoveredData: `{}`, ResourceVersion: 1, CreatedAt: now, UpdatedAt: now}
	if err := db.UpsertCephHost(ctx, &host); err != nil {
		t.Fatal(err)
	}
	status := "online"
	discovered := `{"hostname":"ceph-node-1","address":"10.0.0.10"}`
	if err := db.ReconcileResources(ctx, cluster.ID, 1, []CephEntityRecord{{Kind: "host", NaturalKey: "ceph-node-1", Status: &status, Source: "ceph_cli", ObservedAt: now, DiscoveredData: discovered}}, []string{"host"}); err != nil {
		t.Fatal(err)
	}
	stored, err := db.FindCephHost(ctx, cluster.ID, "ceph-node-1")
	if err != nil {
		t.Fatal(err)
	}
	if stored.SSHAddress != "192.0.2.10" || stored.SSHPort != 2222 || stored.SSHUser != "cephadmin" {
		t.Fatalf("host configuration was overwritten: %#v", stored)
	}
	if stored.DiscoveredData != discovered || stored.Status == nil || *stored.Status != status {
		t.Fatalf("host discovery was not updated: %#v", stored)
	}
	rows, err := db.ListResources(ctx, cluster.ID, ResourceFilter{Kind: "host"})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 {
		t.Fatalf("host resources = %d, want 1", len(rows))
	}
	var configured map[string]any
	if err := json.Unmarshal([]byte(*rows[0].ConfiguredData), &configured); err != nil {
		t.Fatal(err)
	}
	if _, ok := configured["address"]; ok {
		t.Fatalf("empty host address should not override discovered address: %#v", configured)
	}
}

func TestSaveResourceConfigurationMergesEntityConfiguration(t *testing.T) {
	db, err := Open(config.DatabaseConfig{EncryptionKey: schemaTestKey, Engine: EngineSQLite, SQLite: config.SQLiteConfig{Name: "pool-config.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer Close(db)
	ctx := context.Background()
	now := time.Now().UTC()
	cluster := CephCluster{Name: "c", MonitorAddresses: "mon:6789", ClientUsername: "client.test", ClientKey: "cipher", CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(ctx, &cluster); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveResourceConfiguration(ctx, cluster.ID, "pool", "data", `{"name":"data","applications":["rbd"],"compression_mode":"none"}`); err != nil {
		t.Fatal(err)
	}
	if err := db.SaveResourceConfiguration(ctx, cluster.ID, "pool", "data", `{"size":"3"}`); err != nil {
		t.Fatal(err)
	}
	row, err := db.FindResource(ctx, cluster.ID, "pool", "data")
	if err != nil {
		t.Fatal(err)
	}
	var configured map[string]any
	if err := json.Unmarshal([]byte(*row.ConfiguredData), &configured); err != nil {
		t.Fatal(err)
	}
	if configured["name"] != "data" || configured["compression_mode"] != "none" || configured["size"] != "3" {
		t.Fatalf("configured data was not merged: %#v", configured)
	}
	applications, ok := configured["applications"].([]any)
	if !ok || len(applications) != 1 || applications[0] != "rbd" {
		t.Fatalf("applications were not preserved: %#v", configured)
	}
}

func resourceRecord(clusterID uint64, key, name string, status *string, now time.Time, payload map[string]any) CephEntityRecord {
	data, _ := json.Marshal(payload)
	return CephEntityRecord{
		ClusterID: clusterID, Kind: "osd", NaturalKey: key, Name: &name, Status: status,
		Generation: 1, ResourceVersion: 1, Source: "test", ObservedAt: now,
		DiscoveredData: string(data), CreatedAt: now, UpdatedAt: now,
	}
}

func resourceKeys(rows []CephEntityRecord) []string {
	keys := make([]string, 0, len(rows))
	for _, row := range rows {
		keys = append(keys, row.NaturalKey)
	}
	return keys
}

func clusterNames(rows []CephCluster) []string {
	names := make([]string, 0, len(rows))
	for _, row := range rows {
		names = append(names, row.Name)
	}
	return names
}
