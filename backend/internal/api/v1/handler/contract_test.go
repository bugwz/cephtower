package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"cephtower/backend/internal/api/v1/handler"
	"cephtower/backend/internal/api/v1/router"
	"cephtower/backend/internal/config"
	cephprovider "cephtower/backend/internal/integration/ceph"
	authservice "cephtower/backend/internal/service/auth"
	clusterservice "cephtower/backend/internal/service/cluster"
	operationservice "cephtower/backend/internal/service/operation"
	"cephtower/backend/internal/store"
)

const contractKey = "0123456789abcdefghijklmnopqrstuv"

type unusedProvider struct{}

func (unusedProvider) Probe(context.Context, cephprovider.ClusterAccess) (cephprovider.ProbeResult, error) {
	return cephprovider.ProbeResult{}, nil
}

func TestClusterContractEncryptsWriteOnlyKeyAndUsesEnvelope(t *testing.T) {
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: contractKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Path: t.TempDir() + "/api.db"}})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close(db)
	current := func() config.Config {
		return config.Config{Database: config.DatabaseConfig{EncryptionKey: contractKey}}
	}
	auth := authservice.New(func() *store.Database { return db }, current)
	if err := auth.EnsureRoles(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, err := auth.CreateUser(context.Background(), authservice.CreateUserInput{Username: "admin", DisplayName: "Admin", Password: "password-123", Role: "cluster-admin"}); err != nil {
		t.Fatal(err)
	}
	operations := operationservice.New(func() *store.Database { return db }, 1)
	clusters := clusterservice.New(func() *store.Database { return db }, contractKey, operations, unusedProvider{})
	h := handler.New(handler.Dependencies{Auth: auth, Clusters: clusters, Operations: operations, Database: func() *store.Database { return db }})
	mux := http.NewServeMux()
	router.Register(mux, h)
	token := login(t, mux)
	body := `{"name":"test","monitor_addresses":"[v2:192.0.2.1:3300/0,v1:192.0.2.1:6789/0]","client_username":"client.cephtower","client_key":"ceph-secret"}`
	request := httptest.NewRequest(http.MethodPost, "/api/v1/cluster", strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusAccepted {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	assertEnvelope(t, response.Body.Bytes())
	if strings.Contains(response.Body.String(), "ceph-secret") || strings.Contains(response.Body.String(), "client_key") {
		t.Fatalf("secret leaked: %s", response.Body.String())
	}
	if response.Header().Get("Location") == "" {
		t.Fatal("Location header missing")
	}
	rows, err := db.ListClusters(context.Background())
	if err != nil || len(rows) != 1 {
		t.Fatalf("clusters=%v err=%v", rows, err)
	}
	if rows[0].ClientKey == "ceph-secret" || rows[0].ClientKey == "" {
		t.Fatal("client key was not encrypted")
	}
	list := httptest.NewRequest(http.MethodGet, "/api/v1/clusters", nil)
	list.Header.Set("Authorization", "Bearer "+token)
	listResponse := httptest.NewRecorder()
	mux.ServeHTTP(listResponse, list)
	if listResponse.Code != 200 {
		t.Fatalf("list status=%d body=%s", listResponse.Code, listResponse.Body.String())
	}
	assertEnvelope(t, listResponse.Body.Bytes())
	if strings.Contains(listResponse.Body.String(), "client_key") {
		t.Fatalf("client key field leaked: %s", listResponse.Body.String())
	}
}

func TestClusterContractRejectsUnknownAndDuplicateClusterID(t *testing.T) {
	mux, token := contractServer(t)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/cluster", strings.NewReader(`{"name":"x","monitor_addresses":"mon:6789","client_username":"client.x","client_key":"secret","fsid":"forbidden"}`))
	request.Header.Set("Authorization", "Bearer "+token)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != 400 {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	assertEnvelope(t, response.Body.Bytes())
	request = httptest.NewRequest(http.MethodGet, "/api/v1/overview?cluster_id=1", strings.NewReader(`{"cluster_id":1}`))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	response = httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != 400 {
		t.Fatalf("status=%d body=%s", response.Code, response.Body.String())
	}
	assertEnvelope(t, response.Body.Bytes())
}

func contractServer(t *testing.T) (*http.ServeMux, string) {
	t.Helper()
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: contractKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Path: t.TempDir() + "/api.db"}})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	current := func() config.Config {
		return config.Config{Database: config.DatabaseConfig{EncryptionKey: contractKey}}
	}
	auth := authservice.New(func() *store.Database { return db }, current)
	_ = auth.EnsureRoles(context.Background())
	_, _ = auth.CreateUser(context.Background(), authservice.CreateUserInput{Username: "admin", DisplayName: "Admin", Password: "password-123", Role: "cluster-admin"})
	operations := operationservice.New(func() *store.Database { return db }, 1)
	clusters := clusterservice.New(func() *store.Database { return db }, contractKey, operations, unusedProvider{})
	h := handler.New(handler.Dependencies{Auth: auth, Clusters: clusters, Operations: operations, Database: func() *store.Database { return db }})
	mux := http.NewServeMux()
	router.Register(mux, h)
	return mux, login(t, mux)
}

func login(t *testing.T, mux http.Handler) string {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", bytes.NewBufferString(`{"username":"admin","password":"password-123"}`))
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != 200 {
		t.Fatalf("login status=%d body=%s", response.Code, response.Body.String())
	}
	var envelope struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	return envelope.Data.Token
}

func assertEnvelope(t *testing.T, data []byte) {
	t.Helper()
	var value map[string]any
	if err := json.Unmarshal(data, &value); err != nil {
		t.Fatal(err)
	}
	if len(value) != 3 || value["code"] == nil || value["message"] == nil {
		t.Fatalf("invalid envelope: %s", data)
	}
	if _, ok := value["data"]; !ok {
		t.Fatalf("data missing: %s", data)
	}
}
