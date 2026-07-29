package cluster

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"cephtower/backend/internal/config"
	cephprovider "cephtower/backend/internal/integration/ceph"
	"cephtower/backend/internal/security"
	"cephtower/backend/internal/store"
)

const clusterTestKey = "0123456789abcdefghijklmnopqrstuv"

type probeProvider struct{ err error }

func (p probeProvider) Probe(_ context.Context, access cephprovider.ClusterAccess) (cephprovider.ProbeResult, error) {
	if p.err != nil {
		return cephprovider.ProbeResult{}, p.err
	}
	return cephprovider.ProbeResult{FSID: "00000000-0000-0000-0000-000000000001", Version: "ceph version 20.2.2", Capabilities: []cephprovider.Capability{{Name: "ceph_cli", Supported: true}}, Status: map[string]any{"client": access.ClientUsername}}, nil
}

type recordingProbeProvider struct {
	err    error
	called chan cephprovider.ClusterAccess
}

func (p recordingProbeProvider) Probe(_ context.Context, access cephprovider.ClusterAccess) (cephprovider.ProbeResult, error) {
	p.called <- access
	if p.err != nil {
		return cephprovider.ProbeResult{}, p.err
	}
	return cephprovider.ProbeResult{FSID: "00000000-0000-0000-0000-000000000001", Version: "ceph version 20.2.2", Capabilities: []cephprovider.Capability{{Name: "ceph_cli", Supported: true}}, Status: map[string]any{"client": access.ClientUsername}}, nil
}

func clusterTestServices(t *testing.T, provider cephprovider.ClusterProvider) (*Service, *store.Database, store.CephCluster) {
	t.Helper()
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: clusterTestKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Name: "cluster.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	encrypted, err := security.Encrypt([]byte("old-secret"), clusterTestKey)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	row := store.CephCluster{Name: "before", MonitorAddresses: "mon.example.test:6789", ClientUsername: "client.before", ClientKey: encrypted, CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &row); err != nil {
		t.Fatal(err)
	}
	return New(func() *store.Database { return db }, clusterTestKey, provider), db, row
}

func TestUpdatePersistsBeforeAsyncProbe(t *testing.T) {
	called := make(chan cephprovider.ClusterAccess, 1)
	service, db, row := clusterTestServices(t, recordingProbeProvider{called: called})
	name, username, clientKey := "after", "client.after", "new-secret"
	if _, err := service.Update(context.Background(), row.ID, UpdateInput{Name: &name, ClientUsername: &username, ClientKey: &clientKey}); err != nil {
		t.Fatal(err)
	}
	stored, _ := db.FindCluster(context.Background(), row.ID)
	if stored.Name != name || stored.ClientUsername != username {
		t.Fatalf("cluster was not updated: %#v", stored)
	}
	plain, err := security.Decrypt(stored.ClientKey, clusterTestKey)
	if err != nil || string(plain) != clientKey {
		t.Fatalf("client key = %q, err=%v", plain, err)
	}
	select {
	case access := <-called:
		if access.ClientUsername != username {
			t.Fatalf("probe used username %q", access.ClientUsername)
		}
	case <-time.After(time.Second):
		t.Fatal("async probe was not started")
	}
	if _, err := waitForObservationSuccess(t, db, row.ID); err != nil {
		t.Fatal(err)
	}
}

func TestUpdateProbeFailureStillKeepsSavedCandidate(t *testing.T) {
	called := make(chan cephprovider.ClusterAccess, 1)
	service, db, row := clusterTestServices(t, recordingProbeProvider{err: errors.New("unreachable"), called: called})
	name := "after"
	_, err := service.Update(context.Background(), row.ID, UpdateInput{Name: &name})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	stored, _ := db.FindCluster(context.Background(), row.ID)
	if stored.Name != name {
		t.Fatalf("update was not saved: %#v", stored)
	}
	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("async probe was not started")
	}
	observation, err := waitForObservationError(t, db, row.ID)
	if err != nil {
		t.Fatal(err)
	}
	if observation.LastErrorMessage == nil || !strings.Contains(*observation.LastErrorMessage, "unreachable") {
		t.Fatalf("observation error = %#v", observation.LastErrorMessage)
	}
}

func TestCreateProbeFailureKeepsCluster(t *testing.T) {
	called := make(chan cephprovider.ClusterAccess, 1)
	service, db, _ := clusterTestServices(t, recordingProbeProvider{err: errors.New("unreachable"), called: called})
	cluster, result, err := service.Create(context.Background(), CreateInput{Name: "created", MonitorAddresses: "mon2.example.test:6789", ClientUsername: "client.created", ClientKey: "secret"})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if result.ResourceURL == "" || result.Details != nil {
		t.Fatalf("result = %#v", result)
	}
	select {
	case <-called:
	case <-time.After(time.Second):
		t.Fatal("async probe was not started")
	}
	stored, err := db.FindCluster(context.Background(), cluster.ID)
	if err != nil || stored.Name != "created" {
		t.Fatalf("created cluster = %#v, err=%v", stored, err)
	}
	if _, err := waitForObservationError(t, db, cluster.ID); err != nil {
		t.Fatal(err)
	}
}

func waitForObservationError(t *testing.T, db *store.Database, clusterID uint64) (store.CephClusterObservation, error) {
	return waitForObservation(t, db, clusterID, func(observation store.CephClusterObservation) bool {
		return observation.LastErrorMessage != nil
	})
}

func waitForObservationSuccess(t *testing.T, db *store.Database, clusterID uint64) (store.CephClusterObservation, error) {
	return waitForObservation(t, db, clusterID, func(observation store.CephClusterObservation) bool {
		return observation.LastSeenAt != nil
	})
}

func waitForObservation(t *testing.T, db *store.Database, clusterID uint64, ready func(store.CephClusterObservation) bool) (store.CephClusterObservation, error) {
	t.Helper()
	deadline := time.After(time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-deadline:
			return store.CephClusterObservation{}, errors.New("timed out waiting for observation")
		case <-ticker.C:
			observation, err := db.FindObservation(context.Background(), clusterID)
			if err == nil && ready(observation) {
				return observation, nil
			}
		}
	}
}
