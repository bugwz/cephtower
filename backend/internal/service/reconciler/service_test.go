package reconciler

import (
	"context"
	"strings"
	"testing"
	"time"

	"cephtower/backend/internal/config"
	cephdomain "cephtower/backend/internal/domain/ceph"
	cephprovider "cephtower/backend/internal/integration/ceph"
	"cephtower/backend/internal/security"
	clusterservice "cephtower/backend/internal/service/cluster"
	"cephtower/backend/internal/store"
)

const reconcilerTestKey = "0123456789abcdefghijklmnopqrstuv"

type optionalCollectorFake struct {
	t *testing.T
}

type metadataCollectorFake struct {
	results []cephprovider.CollectionResult
	index   int
}

func (f *metadataCollectorFake) Collect(context.Context, cephprovider.ClusterAccess, string) ([]cephprovider.Observation, error) {
	result, err := f.CollectWithMetadata(context.Background(), cephprovider.ClusterAccess{}, "")
	return result.Observations, err
}

func (f *metadataCollectorFake) CollectWithMetadata(context.Context, cephprovider.ClusterAccess, string) (cephprovider.CollectionResult, error) {
	result := f.results[f.index]
	f.index++
	return result, nil
}

func (f optionalCollectorFake) Collect(_ context.Context, access cephprovider.ClusterAccess, module string) ([]cephprovider.Observation, error) {
	if module != "storage" {
		f.t.Fatalf("module = %q", module)
	}
	if access.ClientKey != "plain-ceph-key" {
		f.t.Fatalf("collector did not receive the in-memory Ceph key")
	}
	return []cephprovider.Observation{{
		Kind:       "rgw_user",
		NaturalKey: "fixture-user",
		Name:       "fixture-user",
		Status:     "available",
		Source:     "rgw_admin",
		Payload: map[string]any{
			"uid":        "fixture-user",
			"access_key": "fixture-access-key",
			"keys": []any{map[string]any{
				"secret_key": "fixture-secret-key",
				"token":      "fixture-session-token",
			}},
		},
		ObservedAt: time.Now().UTC(),
	}}, nil
}

func TestOptionalObservationIsStoredWithoutSecrets(t *testing.T) {
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: reconcilerTestKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Name: "reconciler.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close(db)
	encrypted, err := security.Encrypt([]byte("plain-ceph-key"), reconcilerTestKey)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "fixture", MonitorAddresses: "mon:6789", ClientUsername: "client.fixture", ClientKey: encrypted, CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	clusters := clusterservice.New(func() *store.Database { return db }, reconcilerTestKey, nil)
	service := New(func() *store.Database { return db }, clusters, optionalCollectorFake{t: t})
	module := Module{Name: "storage", Kinds: []string{"rgw_user"}}
	if err := service.Reconcile(context.Background(), cluster.ID, module); err != nil {
		t.Fatal(err)
	}
	row, err := db.FindResource(context.Background(), cluster.ID, "rgw_user", "fixture-user")
	if err != nil {
		t.Fatal(err)
	}
	for _, secret := range []string{"plain-ceph-key", "fixture-access-key", "fixture-secret-key", "fixture-session-token"} {
		if strings.Contains(row.DiscoveredData, secret) {
			t.Fatalf("secret %q persisted in %s", secret, row.DiscoveredData)
		}
	}
	if !strings.Contains(row.DiscoveredData, "fixture-user") || !strings.Contains(row.DiscoveredData, "[REDACTED]") {
		t.Fatalf("optional resource discovery was not preserved and redacted: %s", row.DiscoveredData)
	}
	result, err := service.Refresh(context.Background(), cluster.ID, []string{"storage"})
	if err != nil {
		t.Fatal(err)
	}
	modules, ok := result.Details.(map[string]any)["modules"].([]string)
	if !ok || len(modules) != 1 || modules[0] != "storage" {
		t.Fatalf("refresh result = %#v", result)
	}
}

func TestReconcileMarksSuccessfulEmptyKindsStaleButPreservesUnavailableKinds(t *testing.T) {
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: reconcilerTestKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Name: "reconciler-stale.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close(db)
	encrypted, err := security.Encrypt([]byte("plain-ceph-key"), reconcilerTestKey)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "fixture", MonitorAddresses: "mon:6789", ClientUsername: "client.fixture", ClientKey: encrypted, CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	collector := &metadataCollectorFake{results: []cephprovider.CollectionResult{
		{Observations: []cephprovider.Observation{
			{Kind: "rgw_user", NaturalKey: "gone", Payload: map[string]any{"uid": "gone"}, ObservedAt: now},
			{Kind: "rgw_account", NaturalKey: "retained", Payload: map[string]any{"id": "retained"}, ObservedAt: now},
		}},
		{UnavailableKinds: []string{"rgw_account"}},
	}}
	clusters := clusterservice.New(func() *store.Database { return db }, reconcilerTestKey, nil)
	service := New(func() *store.Database { return db }, clusters, collector)
	module := Module{Name: "storage", Kinds: []string{"rgw_user", "rgw_account"}}
	if err := service.Reconcile(context.Background(), cluster.ID, module); err != nil {
		t.Fatal(err)
	}
	if err := service.Reconcile(context.Background(), cluster.ID, module); err != nil {
		t.Fatal(err)
	}
	gone, err := db.FindResource(context.Background(), cluster.ID, "rgw_user", "gone")
	if err != nil || gone.StaleAt == nil {
		t.Fatalf("successful empty kind was not marked stale: row=%#v err=%v", gone, err)
	}
	retained, err := db.FindResource(context.Background(), cluster.ID, "rgw_account", "retained")
	if err != nil || retained.StaleAt != nil {
		t.Fatalf("unavailable optional kind was marked stale: row=%#v err=%v", retained, err)
	}
}

func TestClusterDiscoveryUpdateUsesOverviewAndCoreDaemonVersion(t *testing.T) {
	now := time.Now().UTC()
	version := "ceph version 20.2.2 (0fcffee29411e3a38036764817b6e1afc59741cc) tentacle (stable)"
	update, ok := clusterDiscoveryUpdate(42, []cephprovider.Observation{
		{Kind: "overview", Payload: cephdomain.Overview{FSID: "9f11d824-8e53-11f1-8c55-00163e11e24f", CephVersion: version}, ObservedAt: now},
		{Kind: "daemon", Payload: cephdomain.Daemon{Type: "alertmanager", Version: stringPointer("0.27.0")}, ObservedAt: now},
		{Kind: "daemon", Payload: cephdomain.Daemon{Type: "mgr", Version: stringPointer("20.2.2")}, ObservedAt: now},
	})
	if !ok {
		t.Fatal("cluster discovery update was not detected")
	}
	if update.FSID == nil || *update.FSID != "9f11d824-8e53-11f1-8c55-00163e11e24f" {
		t.Fatalf("fsid = %#v", update.FSID)
	}
	if update.CephVersion == nil || *update.CephVersion != "20.2.2 (0fcffee29411e3a38036764817b6e1afc59741cc)" {
		t.Fatalf("ceph version = %#v", update.CephVersion)
	}
	if update.Status != "available" || update.Generation != 42 {
		t.Fatalf("status/generation = %q/%d", update.Status, update.Generation)
	}
}

func TestReconcileStoresClusterDiscovery(t *testing.T) {
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: reconcilerTestKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Name: "reconciler-observation.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close(db)
	encrypted, err := security.Encrypt([]byte("plain-ceph-key"), reconcilerTestKey)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "fixture", MonitorAddresses: "mon:6789", ClientUsername: "client.fixture", ClientKey: encrypted, CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	collector := &metadataCollectorFake{results: []cephprovider.CollectionResult{{Observations: []cephprovider.Observation{
		{Kind: "overview", NaturalKey: "overview", Payload: cephdomain.Overview{FSID: "9f11d824-8e53-11f1-8c55-00163e11e24f", CephVersion: "ceph version 20.2.2 (0fcffee29411e3a38036764817b6e1afc59741cc) tentacle (stable)"}, ObservedAt: now},
		{Kind: "daemon", NaturalKey: "mgr.a", Payload: cephdomain.Daemon{Type: "mgr", Version: stringPointer("ceph version 20.2.2")}, ObservedAt: now},
	}}}}
	clusters := clusterservice.New(func() *store.Database { return db }, reconcilerTestKey, nil)
	service := New(func() *store.Database { return db }, clusters, collector)
	module := Module{Name: "fast", Kinds: []string{"overview", "daemon"}}
	if err := service.Reconcile(context.Background(), cluster.ID, module); err != nil {
		t.Fatal(err)
	}
	stored, err := db.FindCluster(context.Background(), cluster.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.FSID == nil || *stored.FSID != "9f11d824-8e53-11f1-8c55-00163e11e24f" {
		t.Fatalf("fsid = %#v", stored.FSID)
	}
	if stored.CephVersion == nil || *stored.CephVersion != "20.2.2 (0fcffee29411e3a38036764817b6e1afc59741cc)" {
		t.Fatalf("ceph version = %#v", stored.CephVersion)
	}
	if !strings.Contains(stored.DiscoveredData, *stored.FSID) {
		t.Fatalf("cluster discovery JSON = %s", stored.DiscoveredData)
	}
}

func TestSyncClusterDiscoveryKeepsExistingVersionWithCommit(t *testing.T) {
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: reconcilerTestKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Name: "reconciler-version.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close(db)
	encrypted, err := security.Encrypt([]byte("plain-ceph-key"), reconcilerTestKey)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "fixture", MonitorAddresses: "mon:6789", ClientUsername: "client.fixture", ClientKey: encrypted, CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	version := "20.2.2 (0fcffee29411e3a38036764817b6e1afc59741cc)"
	cluster.CephVersion, cluster.Status, cluster.Enabled, cluster.DiscoveredData = &version, "available", true, `{}`
	if err := db.UpdateClusterDiscovery(context.Background(), cluster); err != nil {
		t.Fatal(err)
	}
	clusters := clusterservice.New(func() *store.Database { return db }, reconcilerTestKey, nil)
	service := New(func() *store.Database { return db }, clusters, &metadataCollectorFake{})
	if err := service.syncClusterDiscovery(context.Background(), cluster.ID, 2, []cephprovider.Observation{
		{Kind: "daemon", Payload: cephdomain.Daemon{Type: "mgr", Version: stringPointer("20.2.2")}, ObservedAt: now},
	}, nil); err != nil {
		t.Fatal(err)
	}
	stored, err := db.FindCluster(context.Background(), cluster.ID)
	if err != nil {
		t.Fatal(err)
	}
	if stored.CephVersion == nil || *stored.CephVersion != version {
		t.Fatalf("ceph version = %#v, want %q", stored.CephVersion, version)
	}
}

func stringPointer(value string) *string {
	return &value
}
