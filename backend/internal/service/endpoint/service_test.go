package endpoint

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/store"
)

const testKey = "0123456789abcdefghijklmnopqrstuv"

func testService(t *testing.T) (*Service, *store.Database, store.CephCluster) {
	t.Helper()
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: testKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Name: "endpoint.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "test", MonitorAddresses: "mon:6789", ClientUsername: "client.test", ClientKey: "encrypted", CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	return New(func() *store.Database { return db }, testKey), db, cluster
}

func TestCredentialIsEncryptedAndDecryptsIntoTypedValue(t *testing.T) {
	service, db, cluster := testService(t)
	plain := "bearer-secret"
	row, err := service.PutCredential(context.Background(), cluster.ID, CredentialInput{Kind: "prometheus", Value: map[string]any{"token": plain}})
	if err != nil {
		t.Fatal(err)
	}
	if row.Credential == plain || strings.Contains(row.Credential, plain) || row.Fingerprint == "" {
		t.Fatalf("credential was not protected: %#v", row)
	}
	stored, err := db.FindCredential(context.Background(), cluster.ID, "prometheus")
	if err != nil {
		t.Fatal(err)
	}
	var decoded struct {
		Token string `json:"token"`
	}
	if err := service.DecryptCredential(context.Background(), cluster.ID, "prometheus", &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded.Token != plain {
		t.Fatalf("decoded=%q stored=%q", decoded.Token, stored.Credential)
	}
	if _, err := json.Marshal(row); err != nil {
		t.Fatal(err)
	}
}

func TestEndpointRejectsCredentialsInURLAndForeignCA(t *testing.T) {
	service, db, cluster := testService(t)
	if _, err := service.CreateEndpoint(context.Background(), cluster.ID, EndpointInput{Kind: "prometheus", URL: "https://user:secret@example.test"}); err == nil {
		t.Fatal("URL credentials accepted")
	}
	now := time.Now().UTC()
	other := store.CephCluster{Name: "other", MonitorAddresses: "mon:6789", ClientUsername: "client.other", ClientKey: "encrypted", CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &other); err != nil {
		t.Fatal(err)
	}
	credential, err := service.PutCredential(context.Background(), other.ID, CredentialInput{Kind: "ca", Value: map[string]any{"ca_certificate": "certificate"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.CreateEndpoint(context.Background(), cluster.ID, EndpointInput{Kind: "prometheus", URL: "https://example.test", TLSMode: "verify_custom_ca", CACredentialID: &credential.ID}); err == nil {
		t.Fatal("foreign CA accepted")
	}
}
