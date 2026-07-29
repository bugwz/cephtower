package reconciler

import (
	"context"
	"strings"
	"testing"
	"time"

	"cephtower/backend/internal/config"
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
		if strings.Contains(row.PayloadJSON, secret) {
			t.Fatalf("secret %q persisted in %s", secret, row.PayloadJSON)
		}
	}
	if !strings.Contains(row.PayloadJSON, "fixture-user") || !strings.Contains(row.PayloadJSON, "[REDACTED]") {
		t.Fatalf("optional resource payload was not preserved and redacted: %s", row.PayloadJSON)
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
