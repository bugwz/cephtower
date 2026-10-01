package router

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

type cephFSCreationTimeExecutor struct{ filesystemFieldsExecutor }

func (e *cephFSCreationTimeExecutor) Run(ctx context.Context, access executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	fixtures := map[string]string{
		"fs subvolumegroup ls enabled --format json":                        `[{"name":"team"}]`,
		"fs subvolumegroup info enabled team --format json":                 `{"created_at":"2020-01-01 01:02:03","bytes_pcent":"25.00"}`,
		"fs subvolume ls enabled --format json":                             `[]`,
		"fs subvolume ls enabled team --format json":                        `[{"name":"volume"},{"name":"missing"},{"name":"retained"},{"name":"clone"}]`,
		"fs subvolume info enabled clone team --format json":                `{"state":"in-progress","type":"clone"}`,
		"fs clone status enabled clone team --format json":                  `{"status":{"state":"in-progress","source":{"volume":"enabled","subvolume":"volume","snapshot":"snap"},"progress_report":{"percentage cloned":"42.5%","amount cloned":"4.2K/10.0K","files cloned":"21/50"}}}`,
		"fs subvolume snapshot ls enabled clone team --format json":         `[]`,
		"fs subvolume info enabled volume team --format json":               `{"state":"complete","type":"subvolume","pool_namespace":"isolated-volume","created_at":"2020-02-01 01:02:03","bytes_pcent":"50.00"}`,
		"fs subvolume info enabled retained team --format json":             `{"state":"snapshot-retained","type":"subvolume","features":["snapshot-retention","snapshot-clone"]}`,
		"fs subvolume snapshot ls enabled retained team --format json":      `[]`,
		"fs subvolume info enabled missing team --format json":              `{"bytes_pcent":"undefined"}`,
		"fs subvolume snapshot ls enabled volume team --format json":        `[{"name":"snap","created_at":"2019-01-01 00:00:00"}]`,
		"fs subvolume snapshot ls enabled missing team --format json":       `[]`,
		"fs subvolume snapshot info enabled volume snap team --format json": `{"created_at":"2020-03-01 01:02:03","has_pending_clones":"yes","pending_clones":[{"name":"clone-default"},{"name":"clone-team","target_group":"team"}],"orphan_clones_count":2}`,
	}
	if output, ok := fixtures[strings.Join(spec.Args, " ")]; ok {
		if spec.Binary != executor.BinaryCeph || spec.Mutating {
			panic("creation time collection must use read-only ceph commands")
		}
		return executor.CommandResult{Stdout: []byte(output)}, nil
	}
	return e.filesystemFieldsExecutor.Run(ctx, access, spec)
}

func TestCephFSCreationTimesFromNativeCollectionToAPI(t *testing.T) {
	const key = "0123456789abcdefghijklmnopqrstuv"
	db, err := store.Open(config.DatabaseConfig{Engine: store.EngineSQLite, EncryptionKey: key, SQLite: config.SQLiteConfig{Name: "cephfs-creation.db"}}, t.TempDir())
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
	provider := &cephprovider.NativeProvider{Executor: &cephFSCreationTimeExecutor{}}
	clusters := clusterservice.New(database, key, provider)
	service := reconciler.New(database, clusters, provider, reconciler.Options{})
	if _, err := service.RefreshKinds(context.Background(), cluster.ID, []string{"subvolume_group", "subvolume", "cephfs_snapshot"}); err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	Register(mux, handler.New(handler.Dependencies{Database: database, Clusters: clusters, AuthEnabled: func() bool { return false }}))
	for _, tt := range []struct{ path, name, created, usage string }{
		{"/filesystem/subvolume/groups", "team", "2020-01-01 01:02:03", "25.00"},
		{"/filesystem/subvolumes", "volume", "2020-02-01 01:02:03", "50.00"},
		{"/filesystem/subvolumes", "missing", "", "undefined"},
		{"/filesystem/subvolumes", "retained", "", ""},
		{"/filesystem/subvolumes", "clone", "", ""},
		{"/filesystem/subvolume/snapshots", "snap", "2020-03-01 01:02:03", ""},
	} {
		body, _ := json.Marshal(map[string]any{"cluster_id": cluster.ID, "fs": "enabled", "name": tt.name})
		req := httptest.NewRequest("GET", "/api/v1"+tt.path, strings.NewReader(string(body)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		var result struct {
			Data struct {
				Items []struct {
					Name      string         `json:"name"`
					Status    string         `json:"status"`
					CreatedAt string         `json:"created_at"`
					Data      map[string]any `json:"data"`
				} `json:"items"`
			} `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &result); err != nil || rec.Code != 200 {
			t.Fatalf("GET %s: %d %s", tt.path, rec.Code, rec.Body.String())
		}
		found := false
		for _, row := range result.Data.Items {
			if row.Name != tt.name {
				continue
			}
			found = true
			value, exists := row.Data["ceph_created_at"]
			if (tt.created == "" && exists) || (tt.created != "" && value != tt.created) {
				t.Fatalf("native time for %s = %v", tt.name, value)
			}
			if _, exists := row.Data["created_at"]; exists {
				t.Fatal("native time collides with cache timestamp")
			}
			if row.CreatedAt == "" || row.CreatedAt == tt.created {
				t.Fatal("cache creation metadata was replaced")
			}
			if tt.usage != "" && row.Data["bytes_pcent"] != tt.usage {
				t.Fatalf("usage=%v", row.Data["bytes_pcent"])
			}
			if tt.path == "/filesystem/subvolume/snapshots" {
				clones, ok := row.Data["pending_clones"].([]any)
				if !ok || len(clones) != 2 || row.Data["has_pending_clones"] != "yes" || row.Data["orphan_clones_count"] != float64(2) {
					t.Fatalf("snapshot dependencies=%v", row.Data)
				}
				first := clones[0].(map[string]any)
				second := clones[1].(map[string]any)
				if first["name"] != "clone-default" || second["name"] != "clone-team" || second["target_group"] != "team" {
					t.Fatalf("native clone scope=%v", clones)
				}
				if _, exists := first["target_group"]; exists {
					t.Fatal("default group must remain native omission in API")
				}
			}
			if tt.path == "/filesystem/subvolumes" {
				wantState := map[string]string{"volume": "complete", "missing": "unknown", "retained": "snapshot-retained", "clone": "in-progress"}[tt.name]
				if row.Status != wantState {
					t.Fatalf("subvolume %s status=%s want=%s", tt.name, row.Status, wantState)
				}
				wantType := "subvolume"
				if tt.name == "clone" {
					wantType = "clone"
				}
				if tt.name != "missing" && (row.Data["state"] != wantState || row.Data["type"] != wantType) {
					t.Fatalf("native state/type=%v", row.Data)
				}
				if tt.name == "volume" && row.Data["pool_namespace"] != "isolated-volume" {
					t.Fatal("isolated namespace missing")
				}
				if tt.name == "retained" {
					if _, exists := row.Data["path"]; exists {
						t.Fatal("retained subvolume has fabricated path")
					}
				}
				if tt.name == "clone" {
					progress, ok := row.Data["clone_progress"].(map[string]any)
					if !ok || progress["percentage cloned"] != "42.5%" || progress["amount cloned"] != "4.2K/10.0K" || progress["files cloned"] != "21/50" || row.Data["clone_state"] != "in-progress" {
						t.Fatalf("clone report=%v", row.Data)
					}
				}
			}
		}
		if !found {
			t.Fatalf("resource %s missing in %s", tt.name, rec.Body.String())
		}
	}
}
