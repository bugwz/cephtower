package router

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"cephtower/backend/internal/api/v1/handler"
	"cephtower/backend/internal/config"
	cephprovider "cephtower/backend/internal/integration/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/security"
	clusterservice "cephtower/backend/internal/service/cluster"
	"cephtower/backend/internal/service/reconciler"
	"cephtower/backend/internal/store"
)

type filesystemFieldsExecutor struct {
	specs   []executor.CommandSpec
	invalid bool
}

func (e *filesystemFieldsExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	output := ""
	switch spec.ID {
	case "collect.osd_tree":
		output = `{"nodes":[]}`
	case "collect.osd_dump":
		output = `{"osds":[]}`
	case "collect.pool":
		output = `[]`
	case "collect.fs":
		output = `{"filesystems":[{"mdsmap":{"fs_name":"enabled","id":1,"enabled":true,"created":"2026-09-25T10:20:30.123456+0000"}},{"mdsmap":{"fs_name":"disabled","id":2,"enabled":false,"created":"2026-09-24 10:20:30"}},{"mdsmap":{"fs_name":"unknown","id":3}}]}`
		if e.invalid {
			output = `{"filesystems":[{"mdsmap":{"fs_name":"enabled","enabled":"false"}}]}`
		}
	default:
		return executor.CommandResult{}, errors.New("fixture command unavailable")
	}
	return executor.CommandResult{Stdout: []byte(output)}, nil
}

func TestFilesystemEnabledAndCreatedFromNativeCollectionToAPI(t *testing.T) {
	const key = "0123456789abcdefghijklmnopqrstuv"
	db, err := store.Open(config.DatabaseConfig{Engine: store.EngineSQLite, EncryptionKey: key, SQLite: config.SQLiteConfig{Name: "filesystem-fields.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	secret, _ := security.Encrypt([]byte("fixture"), key)
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "fixture", MonitorAddresses: "mon:6789", ClientUsername: "client.admin", ClientKey: secret, CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	if err := db.UpsertCapabilities(context.Background(), []store.CephClusterCapability{{ClusterID: cluster.ID, Name: "cephfs_volume", Supported: true, ObservedAt: now, UpdatedAt: now}}); err != nil {
		t.Fatal(err)
	}
	database := func() *store.Database { return db }
	runner := &filesystemFieldsExecutor{}
	provider := &cephprovider.NativeProvider{Executor: runner}
	clusters := clusterservice.New(database, key, provider)
	service := reconciler.New(database, clusters, provider, reconciler.Options{})
	if _, err := service.RefreshKinds(context.Background(), cluster.ID, []string{"filesystem"}); err != nil {
		t.Fatal(err)
	}
	found := false
	for _, spec := range runner.specs {
		if spec.ID == "collect.fs" {
			found = true
			if spec.Mutating || spec.Binary != executor.BinaryCeph || !reflect.DeepEqual(spec.Args, []string{"fs", "dump", "--format", "json"}) {
				t.Fatalf("native command=%+v", spec)
			}
		}
	}
	if !found {
		t.Fatal("filesystem map was not collected")
	}
	mux := http.NewServeMux()
	Register(mux, handler.New(handler.Dependencies{Database: database, Clusters: clusters, AuthEnabled: func() bool { return false }}))
	for _, tt := range []struct {
		name, created string
		enabled       *bool
	}{
		{"enabled", "2026-09-25T10:20:30.123456+0000", boolPointer(true)},
		{"disabled", "2026-09-24 10:20:30", boolPointer(false)},
		{"unknown", "", nil},
	} {
		body, _ := json.Marshal(map[string]any{"cluster_id": cluster.ID, "name": tt.name})
		req := httptest.NewRequest("GET", "/api/v1/filesystem", strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		var result struct {
			Data struct {
				CreatedAt string         `json:"created_at"`
				Data      map[string]any `json:"data"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil || rec.Code != 200 {
			t.Fatalf("GET %s: %d %s", tt.name, rec.Code, rec.Body.String())
		}
		value, exists := result.Data.Data["enabled"]
		if !exists || (tt.enabled == nil && value != nil) || (tt.enabled != nil && value != *tt.enabled) {
			t.Fatalf("enabled=%v expected=%v", value, tt.enabled)
		}
		created, exists := result.Data.Data["created"]
		if !exists || (tt.created == "" && created != nil) || (tt.created != "" && created != tt.created) {
			t.Fatalf("created=%v expected=%s", created, tt.created)
		}
		if result.Data.CreatedAt == "" || result.Data.CreatedAt == tt.created {
			t.Fatal("native and cache creation timestamps were mixed")
		}
	}
	runner.invalid = true
	if _, err := service.RefreshKinds(context.Background(), cluster.ID, []string{"filesystem"}); err == nil {
		t.Fatal("string enabled value was silently coerced")
	}
}

func boolPointer(value bool) *bool { return &value }
