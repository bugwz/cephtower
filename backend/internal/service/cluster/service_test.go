package cluster

import (
	"context"
	"testing"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/store"
)

func TestCreateValidatesBeforePersisting(t *testing.T) {
	db, err := store.Open(config.DatabaseConfig{Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Path: t.TempDir() + "/cluster.db"}})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close(db)
	service := New(Dependencies{Database: func() *store.Database { return db }})
	if err := service.Create(context.Background(), Input{Name: "primary"}); err == nil {
		t.Fatal("Create() succeeded with missing connection fields")
	}
	count, err := db.CountRecords(context.Background(), &store.CephCluster{}, nil)
	if err != nil || count != 0 {
		t.Fatalf("cluster count = %d, error = %v", count, err)
	}
}
