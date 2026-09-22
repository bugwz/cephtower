package store

import (
	"context"
	"testing"
	"time"

	"cephtower/backend/internal/config"
)

func TestObservationHistorySamplingAndRetention(t *testing.T) {
	db, err := Open(config.DatabaseConfig{EncryptionKey: schemaTestKey, Engine: EngineSQLite, SQLite: config.SQLiteConfig{Name: "history.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = Close(db) })
	cluster := CephCluster{Name: "history", MonitorAddresses: "mon:6789", ClientUsername: "client.test", ClientKey: "cipher"}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	base := time.Now().UTC().Add(-2 * time.Hour).Truncate(time.Second)
	for index, offset := range []time.Duration{0, time.Minute, 5 * time.Minute} {
		row := CephObservationHistory{ClusterID: cluster.ID, Kind: "overview", NaturalKey: "overview", Source: "ceph_cli", ObservedAt: base.Add(offset), DataJSON: `{}`}
		inserted, err := db.AppendObservationHistoryIfDue(context.Background(), &row, 5*time.Minute)
		if err != nil {
			t.Fatal(err)
		}
		if inserted != (index != 1) {
			t.Fatalf("sample %d inserted=%v", index, inserted)
		}
	}
	rows, err := db.ListObservationHistory(context.Background(), ObservationHistoryFilter{ClusterID: cluster.ID, Kind: "overview", NaturalKey: "overview"})
	if err != nil || len(rows) != 2 || !rows[0].ObservedAt.Equal(base.Add(5*time.Minute)) {
		t.Fatalf("history=%#v err=%v", rows, err)
	}
	deleted, err := db.PruneObservationHistory(context.Background(), "overview", base.Add(time.Minute))
	if err != nil || deleted != 1 {
		t.Fatalf("pruned=%d err=%v", deleted, err)
	}
}
