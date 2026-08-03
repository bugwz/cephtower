package hostprofile

import (
	"context"
	"reflect"
	"testing"
	"time"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/store"
)

const testEncryptionKey = "0123456789abcdefghijklmnopqrstuv"

func TestSaveSyncsSSHSettingsToSelectedHosts(t *testing.T) {
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: testEncryptionKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Name: "hostprofile.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close(db)

	ctx := context.Background()
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "fixture", MonitorAddresses: "mon:6789", ClientUsername: "client.fixture", ClientKey: "encrypted", CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(ctx, &cluster); err != nil {
		t.Fatal(err)
	}
	target := store.CephHost{
		ClusterID:      cluster.ID,
		Hostname:       "ceph-node-2",
		SSHPort:        22,
		SSHUser:        "root",
		DiscoveredData: `{"hostname":"ceph-node-2","address":"10.0.0.2"}`,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	if err := db.UpsertCephHost(ctx, &target); err != nil {
		t.Fatal(err)
	}

	password := "new-password"
	service := New(func() *store.Database { return db }, testEncryptionKey)
	view, err := service.Save(ctx, SaveInput{
		ClusterID:     cluster.ID,
		Hostname:      "ceph-node-1",
		SSHAddress:    "10.0.0.1",
		SSHPort:       2022,
		SSHUser:       "admin",
		SSHPassword:   &password,
		SyncHostnames: []string{"ceph-node-2"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(view.SyncedHostnames, []string{"ceph-node-2"}) {
		t.Fatalf("synced hostnames = %v, want ceph-node-2", view.SyncedHostnames)
	}

	synced, err := db.FindCephHost(ctx, cluster.ID, "ceph-node-2")
	if err != nil {
		t.Fatal(err)
	}
	if synced.SSHAddress != "10.0.0.2" {
		t.Fatalf("synced host address = %q, want target address", synced.SSHAddress)
	}
	if synced.SSHUser != "admin" || synced.SSHPort != 2022 {
		t.Fatalf("synced SSH settings were not updated: %#v", synced)
	}
	syncedView, err := service.Get(ctx, cluster.ID, "ceph-node-2")
	if err != nil {
		t.Fatal(err)
	}
	if !syncedView.SSHPasswordSet {
		t.Fatal("synced host password was not saved")
	}
}
