package router

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
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
	"cephtower/backend/internal/service/clusterinspect"
	mutationservice "cephtower/backend/internal/service/mutation"
	operationservice "cephtower/backend/internal/service/operation"
	"cephtower/backend/internal/service/reconciler"
	"cephtower/backend/internal/store"
)

type authRouteExecutor struct{ specs []executor.CommandSpec }

func (e *authRouteExecutor) Run(_ context.Context, _ executor.ClusterAccess, spec executor.CommandSpec) (executor.CommandResult, error) {
	e.specs = append(e.specs, spec)
	switch spec.ID {
	case "rgw_zone.sync.status":
		return executor.CommandResult{Stdout: []byte("          realm realm-id (realm-a)\n      zonegroup group-id (group-a)\n           zone zone-id (zone-a)\n  metadata sync no sync (zone is master)\n"), Stderr: []byte("private-sync-diagnostic")}, nil
	case "rgw_realm.token.read":
		token := base64.StdEncoding.EncodeToString([]byte(`{"realm_id":"realm-id","realm_name":"realm-a","endpoint":"https://rgw.example","access_key":"realm-access","secret":"realm-secret"}`))
		raw, _ := json.Marshal([]map[string]string{{"realm": "realm-a", "token": token}})
		return executor.CommandResult{Stdout: raw}, nil
	case "erasure_code.manager":
		return executor.CommandResult{Stdout: []byte(`{"active_name":"node.a"}`)}, nil
	case "erasure_code.configuration":
		return executor.CommandResult{Stdout: []byte(`[{"name":"osd_erasure_code_plugins","value":"isa lrc"},{"name":"erasure_code_dir","value":"/usr/lib/ceph/erasure-code"}]`)}, nil
	case "host.hardware":
		return executor.CommandResult{Stdout: []byte(`{"node1":{"sys":{"dimm1":{"status":{"health":"OK"}}}}}`)}, nil
	case "daemon.perf.schema":
		return executor.CommandResult{Stdout: []byte(`{"osd":{"counter":{"description":"counter","type":10,"units":"bytes"}}}`)}, nil
	case "daemon.perf.dump":
		return executor.CommandResult{Stdout: []byte(`{"osd":{"counter":18446744073709551615}}`)}, nil
	case "telemetry.status":
		return executor.CommandResult{Stdout: []byte(`{"enabled":false,"channel_basic":true,"interval":24,"last_upload":null}`)}, nil
	case "telemetry.report":
		return executor.CommandResult{Stdout: []byte(`{"report":{"counter":9007199254740993}}`)}, nil
	case "telemetry.update.post_check":
		return executor.CommandResult{Stdout: []byte(`{"enabled":true}`)}, nil
	case "cluster.logs":
		return executor.CommandResult{Stdout: []byte(`[{"name":"mon.a","rank":"0","stamp":"2026-09-14 01:00:00","seq":1,"channel":"audit","priority":"[INF]","message":"entry"}]`)}, nil
	case "configuration.help":
		return executor.CommandResult{Stdout: []byte(`{"name":"osd_memory_target","type":"size","default":"4G","can_update_at_runtime":true}`)}, nil
	case "cephfs.directory.list":
		return executor.CommandResult{Stdout: []byte("drwxr-xr-x 4096 1000 1000 2026-09-23 12:00:00 projects/\n")}, nil
	case "cephfs.performance.map":
		return executor.CommandResult{Stdout: []byte(`{"mdsmap":{"fs_name":"cephfs","info":{"gid_1":{"name":"a","gid":1,"rank":0,"state":"up:active"}}}}`)}, nil
	case "cephfs.performance.dump":
		return executor.CommandResult{Stdout: []byte(`{"mds_mem":{"ino":123},"mds_server":{"handle_client_request":9007199254740993}}`)}, nil
	case "cephfs.pools.map":
		return executor.CommandResult{Stdout: []byte(`{"mdsmap":{"fs_name":"cephfs","metadata_pool":1,"data_pools":[2]}}`)}, nil
	case "cephfs.pools.df":
		return executor.CommandResult{Stdout: []byte(`{"pools":[{"id":1,"name":"cephfs.meta","stats":{"stored":50,"max_avail":100,"bytes_used":150}},{"id":2,"name":"cephfs.data","stats":{"stored":9007199254740993,"max_avail":100,"bytes_used":18014398509481986}}]}`)}, nil
	case "cephfs.mds.map":
		return executor.CommandResult{Stdout: []byte(`{"filesystems":[{"mdsmap":{"fs_name":"cephfs","in":[0,1],"up":{"mds_0":1},"info":{"gid_1":{"name":"a","gid":1,"rank":0,"state":"up:active"}}}}],"standbys":[]}`)}, nil
	case "cephfs.mds.metadata":
		return executor.CommandResult{Stdout: []byte(`[{"name":"a","ceph_version":"ceph fixture"}]`)}, nil
	case "cephfs.directory.quota":
		return executor.CommandResult{Stdout: []byte("max_bytes: 1048576\nmax_files: 100\n")}, nil
	case "cephfs.snapshot.list":
		return executor.CommandResult{Stdout: []byte("drwxr-xr-x 0 0 0 2026-09-23 13:00:00 release-one/\n")}, nil
	case "cephfs_entry_snapshot.create.post_check":
		return executor.CommandResult{Stdout: []byte("drwxr-xr-x 0 0 0 2026-09-23 13:00:00 release-one/\n")}, nil
	case "cephfs_entry_snapshot.delete.post_check":
		return executor.CommandResult{}, nil
	case "cephfs_entry.create.post_check":
		return executor.CommandResult{Stdout: []byte("drwxr-xr-x 0 0 0 2026-10-01 13:00:00 new directory/\n")}, nil
	case "cephfs_entry.delete.post_check":
		return executor.CommandResult{}, nil
	case "cephfs_entry.rename.pre_check":
		return executor.CommandResult{Stdout: []byte("drwxr-xr-x 0 0 0 2026-10-01 13:00:00 new directory/\n")}, nil
	case "cephfs_entry.rename.destination_post_check":
		return executor.CommandResult{Stdout: []byte("drwxr-xr-x 0 0 0 2026-10-01 13:00:00 renamed directory/\n")}, nil
	case "cephfs_entry.rename.path_post_check":
		return executor.CommandResult{}, nil
	case "collect.ceph_user":
		return executor.CommandResult{Stdout: []byte(`{"auth_dump":[{"entity":"client.backup","key":"sensitive-fixture-key","caps":{"mon":"allow r"}}]}`)}, nil
	case "collect.config", "config_value.delete.post_check":
		return executor.CommandResult{Stdout: []byte(`[]`)}, nil
	case "ceph_user.export":
		return executor.CommandResult{Stdout: []byte("[client.backup]\n key = sensitive-fixture-key\n")}, nil
	default:
		return executor.CommandResult{}, nil
	}
}

func TestCephUserAPIEndToEndWithoutCluster(t *testing.T) {
	const encryptionKey = "0123456789abcdefghijklmnopqrstuv"
	db, err := store.Open(config.DatabaseConfig{Engine: store.EngineSQLite, EncryptionKey: encryptionKey, SQLite: config.SQLiteConfig{Name: "auth-route.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close(db) })
	key, err := security.Encrypt([]byte("fixture"), encryptionKey)
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	cluster := store.CephCluster{Name: "fixture", MonitorAddresses: "mon:6789", ClientUsername: "client.admin", ClientKey: key, CreatedAt: now, UpdatedAt: now}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatal(err)
	}
	if err := db.UpsertCapabilities(context.Background(), []store.CephClusterCapability{{ClusterID: cluster.ID, Name: "cephfs_data_access", Supported: true, ObservedAt: now, UpdatedAt: now}}); err != nil {
		t.Fatal(err)
	}
	runner := &authRouteExecutor{}
	provider := &cephprovider.NativeProvider{Executor: runner}
	database := func() *store.Database { return db }
	clusters := clusterservice.New(database, encryptionKey, provider)
	mutations := mutationservice.New(clusters, runner)
	reconcileService := reconciler.New(database, clusters, provider, reconciler.Options{})
	operations := operationservice.New(database, encryptionKey, operationservice.NewActionDispatcher(mutations, nil, reconcileService), operationservice.Options{Workers: 1, PollInterval: time.Millisecond})
	if err := operations.Start(context.Background()); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(operations.Stop)
	h := handler.New(handler.Dependencies{Inspection: clusterinspect.New(clusters, runner), Database: database, Clusters: clusters, Mutations: mutations, Operations: operations, Reconciler: reconcileService, AuthEnabled: func() bool { return false }})
	mux := http.NewServeMux()
	Register(mux, h)
	send := func(method, path string, body map[string]any) *httptest.ResponseRecorder {
		t.Helper()
		body["cluster_id"] = cluster.ID
		encoded, _ := json.Marshal(body)
		req := httptest.NewRequest(method, "/api/v1"+path, bytes.NewReader(encoded))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK && rec.Code != http.StatusAccepted {
			t.Fatalf("%s %s: %d %s", method, path, rec.Code, rec.Body.String())
		}
		if rec.Code == http.StatusAccepted {
			var response struct {
				Data struct {
					OperationID uint64 `json:"operation_id"`
				} `json:"data"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil || response.Data.OperationID == 0 {
				t.Fatalf("decode accepted operation: id=%d err=%v", response.Data.OperationID, err)
			}
			deadline := time.Now().Add(time.Second)
			for time.Now().Before(deadline) {
				row, err := db.FindOperation(context.Background(), response.Data.OperationID)
				if err == nil && row.Status == store.OperationSucceeded {
					return rec
				}
				if err == nil && row.Status == store.OperationFailed {
					t.Fatalf("%s %s operation failed: %v %v", method, path, row.ErrorCode, row.ErrorMessage)
				}
				time.Sleep(time.Millisecond)
			}
			t.Fatalf("%s %s operation did not finish", method, path)
		}
		return rec
	}
	perfResult := send("GET", "/daemon/perf", map[string]any{"name": "osd.1"})
	ecInfo := send("GET", "/erasure/code/info", map[string]any{})
	if ecInfo.Header().Get("Cache-Control") != "no-store" || !strings.Contains(ecInfo.Body.String(), `"manager":"mgr.node.a"`) || !strings.Contains(ecInfo.Body.String(), `"plugins":["isa","lrc"]`) {
		t.Fatalf("erasure code configuration response incomplete: %s", ecInfo.Body.String())
	}
	hardwareResult := send("GET", "/host/hardware", map[string]any{"host": "node1", "category": "memory"})
	clusterHardware := send("GET", "/host/hardware", map[string]any{"category": "memory"})
	if clusterHardware.Header().Get("Cache-Control") != "no-store" || !strings.Contains(clusterHardware.Body.String(), `"host":""`) || !strings.Contains(clusterHardware.Body.String(), `"host":"node1"`) {
		t.Fatalf("cluster hardware response lost scope or component host: %s", clusterHardware.Body.String())
	}
	if hardwareResult.Header().Get("Cache-Control") != "no-store" || !strings.Contains(hardwareResult.Body.String(), `"health":"OK"`) || !strings.Contains(hardwareResult.Body.String(), `"host":"node1"`) {
		t.Fatalf("hardware response is incomplete: %s", hardwareResult.Body.String())
	}
	if perfResult.Header().Get("Cache-Control") != "no-store" || !strings.Contains(perfResult.Body.String(), `"daemon_name":"osd.1"`) || !strings.Contains(perfResult.Body.String(), `"raw_value":"18446744073709551615"`) {
		t.Fatalf("daemon performance snapshot lost identity or precision: %s", perfResult.Body.String())
	}
	logResult := send("GET", "/logs", map[string]any{"channel": "audit", "level": "debug", "limit": 30})
	if !strings.Contains(logResult.Body.String(), "entry") {
		t.Fatal("log API did not return Ceph records")
	}
	helpResult := send("GET", "/configuration/option", map[string]any{"name": "osd_memory_target"})
	telemetryResult := send("GET", "/manager/telemetry/status", map[string]any{})
	telemetryReport := send("GET", "/manager/telemetry/report", map[string]any{"mode": "preview"})
	send("PATCH", "/manager/telemetry", map[string]any{"enabled": true, "license": "sharing-1-0"})
	if !strings.Contains(telemetryReport.Body.String(), "9007199254740993") || telemetryReport.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("telemetry report API lost exact uncached report")
	}
	if !strings.Contains(telemetryResult.Body.String(), `"enabled":false`) || telemetryResult.Header().Get("Cache-Control") != "no-store" {
		t.Fatal("telemetry status API did not return uncached native state")
	}
	if !strings.Contains(helpResult.Body.String(), "can_update_at_runtime") {
		t.Fatal("configuration metadata missing")
	}
	directoryResult := send("GET", "/filesystem/entries", map[string]any{"fs": "cephfs", "path": "/"})
	performanceResult := send("GET", "/filesystem/performance", map[string]any{"fs": "cephfs"})
	poolResult := send("GET", "/filesystem/pools", map[string]any{"fs": "cephfs"})
	mdsResult := send("GET", "/filesystem/mds", map[string]any{"fs": "cephfs"})
	missingBody, _ := json.Marshal(map[string]any{"cluster_id": cluster.ID, "fs": "absent"})
	missingReq := httptest.NewRequest("GET", "/api/v1/filesystem/mds", bytes.NewReader(missingBody))
	missingReq.Header.Set("Content-Type", "application/json")
	missingResult := httptest.NewRecorder()
	mux.ServeHTTP(missingResult, missingReq)
	if missingResult.Code != http.StatusNotFound {
		t.Fatalf("missing filesystem returned %d: %s", missingResult.Code, missingResult.Body.String())
	}
	if mdsResult.Header().Get("Cache-Control") != "no-store" || !strings.Contains(mdsResult.Body.String(), `"rank":"1","name":"","gid":"","state":"failed"`) || !strings.Contains(mdsResult.Body.String(), `"version":"ceph fixture"`) {
		t.Fatalf("MDS topology response is incomplete: %s", mdsResult.Body.String())
	}
	if poolResult.Header().Get("Cache-Control") != "no-store" || !strings.Contains(poolResult.Body.String(), `"stored":"9007199254740993"`) || !strings.Contains(poolResult.Body.String(), `"size":"150"`) || !strings.Contains(poolResult.Body.String(), `"type":"metadata"`) {
		t.Fatalf("filesystem pool usage response is incomplete: %s", poolResult.Body.String())
	}
	if performanceResult.Header().Get("Cache-Control") != "no-store" || !strings.Contains(performanceResult.Body.String(), `"value":"9007199254740993"`) || !strings.Contains(performanceResult.Body.String(), `"name":"mds_log.ev","value":null`) {
		t.Fatalf("performance response is incomplete: %s", performanceResult.Body.String())
	}
	if directoryResult.Header().Get("Cache-Control") != "no-store" || !strings.Contains(directoryResult.Body.String(), `"path":"/projects"`) || !strings.Contains(directoryResult.Body.String(), `"max_files":100`) {
		t.Fatalf("directory browser response is incomplete: %s", directoryResult.Body.String())
	}
	snapshotResult := send("GET", "/filesystem/entry/snapshots", map[string]any{"fs": "cephfs", "path": "/projects"})
	if snapshotResult.Header().Get("Cache-Control") != "no-store" || !strings.Contains(snapshotResult.Body.String(), `"name":"release-one"`) {
		t.Fatalf("directory snapshot response is incomplete: %s", snapshotResult.Body.String())
	}
	send("POST", "/filesystem/entry/snapshot", map[string]any{"fs": "cephfs", "path": "/projects", "name": "release-one"})
	send("DELETE", "/filesystem/entry/snapshot", map[string]any{"fs": "cephfs", "path": "/projects", "name": "release-one"})
	send("POST", "/filesystem/entry", map[string]any{"fs": "cephfs", "path": "/projects/new directory"})
	send("DELETE", "/filesystem/entry", map[string]any{"fs": "cephfs", "path": "/projects/new directory"})
	send("PATCH", "/filesystem/entry", map[string]any{"fs": "cephfs", "path": "/projects/new directory", "destination": "/archive/renamed directory"})
	var renameCommands []executor.CommandSpec
	for _, spec := range runner.specs {
		if strings.HasPrefix(spec.ID, "cephfs_entry.rename") {
			renameCommands = append(renameCommands, spec)
		}
	}
	if len(renameCommands) != 4 || renameCommands[0].Mutating || !renameCommands[1].Mutating || renameCommands[2].Mutating || renameCommands[3].Mutating || !reflect.DeepEqual(renameCommands[1].Args, []string{"--fs", "cephfs", "mv", `"/projects/new directory"`, `"/archive/renamed directory"`}) || !reflect.DeepEqual(renameCommands[2].Args, []string{"--fs", "cephfs", "ls", "-la", "/archive"}) || !reflect.DeepEqual(renameCommands[3].Args, []string{"--fs", "cephfs", "ls", "-la", "/projects"}) {
		t.Fatalf("unexpected directory rename chain: %+v", renameCommands)
	}
	send("PUT", "/configuration/value", map[string]any{"who": "osd/host:node-a", "name": "osd_memory_target", "value": "4G"})
	if _, err := db.FindResource(context.Background(), cluster.ID, "config_value", "osd/host:node-a:osd_memory_target"); !errors.Is(err, store.ErrRecordNotFound) {
		t.Fatalf("configuration cache must reflect the empty Ceph response: %v", err)
	}
	send("DELETE", "/configuration/value", map[string]any{"who": "osd/host:node-a", "name": "osd_memory_target"})
	send("PUT", "/configuration/value", map[string]any{"who": "client.rgw", "name": "rgw_keystone_admin_password", "value": "sensitive-config-value"})
	if _, err := db.FindResource(context.Background(), cluster.ID, "config_value", "client.rgw:rgw_keystone_admin_password"); !errors.Is(err, store.ErrRecordNotFound) {
		t.Fatal("configuration request body entered the observed-state cache")
	}

	send("POST", "/resource/refresh", map[string]any{"kind": "ceph_user"})
	rec := send("GET", "/ceph/users", map[string]any{})
	if strings.Contains(rec.Body.String(), "sensitive-fixture-key") || !strings.Contains(rec.Body.String(), "allow r") {
		t.Fatal("list leaked secret or lost caps")
	}
	send("POST", "/ceph/user", map[string]any{"entity": "client.new", "caps": map[string]any{"mon": "allow r"}})
	send("PATCH", "/ceph/user", map[string]any{"entity": "client.backup", "caps": map[string]any{"osd": "profile rbd pool=data"}})
	rec = send("GET", "/ceph/users/export", map[string]any{"entities": []string{"client.backup"}})
	if rec.Header().Get("Cache-Control") != "no-store" || !strings.Contains(rec.Body.String(), "sensitive-fixture-key") {
		t.Fatal("export did not return a no-store keyring")
	}
	send("POST", "/ceph/users/import", map[string]any{"keyring": "[client.imported]\nkey = sensitive-import-key\n"})
	syncResponse := send("POST", "/rgw/zone/sync/status", map[string]any{"zone_id": "zone-id", "name": "zone-a"})
	if !strings.Contains(syncResponse.Body.String(), `"diagnostics_present":true`) || strings.Contains(syncResponse.Body.String(), "private-sync-diagnostic") {
		t.Fatal("sync diagnostics lost or exposed")
	}
	if syncResponse.Code != http.StatusOK || syncResponse.Header().Get("Cache-Control") != "no-store" || strings.Contains(syncResponse.Body.String(), "operation_id") || !strings.Contains(syncResponse.Body.String(), "metadata sync") {
		t.Fatal("sync status was queued, cached or lost")
	}
	for _, body := range []string{`{"cluster_id":0,"zone_id":"zone-id","name":"zone-a"}`, fmt.Sprintf(`{"cluster_id":%d,"zone_id":"wrong","name":"zone-a"}`, cluster.ID), fmt.Sprintf(`{"cluster_id":%d,"zone_id":"zone-id","name":"zone-a","extra":true}`, cluster.ID)} {
		bad := httptest.NewRecorder()
		mux.ServeHTTP(bad, httptest.NewRequest("POST", "/api/v1/rgw/zone/sync/status", strings.NewReader(body)))
		if bad.Code < 400 || bad.Header().Get("Cache-Control") != "no-store" || strings.Contains(bad.Body.String(), "metadata sync") {
			t.Fatal("invalid sync scope accepted")
		}
	}
	realmResponse := send("POST", "/rgw/realm/token", map[string]any{"realm_id": "realm-id", "name": "realm-a"})
	if realmResponse.Code != http.StatusOK || realmResponse.Header().Get("Cache-Control") != "no-store" || strings.Contains(realmResponse.Body.String(), "operation_id") {
		t.Fatal("realm token was cached or queued")
	}
	var realmEnvelope struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	if json.Unmarshal(realmResponse.Body.Bytes(), &realmEnvelope) != nil || realmEnvelope.Data.Token == "" {
		t.Fatal("realm token missing")
	}
	for _, body := range []string{`{"cluster_id":0,"realm_id":"realm-id","name":"realm-a"}`, fmt.Sprintf(`{"cluster_id":%d,"realm_id":"wrong","name":"realm-a"}`, cluster.ID), fmt.Sprintf(`{"cluster_id":%d,"realm_id":"realm-id","name":"realm-a","extra":true}`, cluster.ID)} {
		bad := httptest.NewRecorder()
		mux.ServeHTTP(bad, httptest.NewRequest("POST", "/api/v1/rgw/realm/token", strings.NewReader(body)))
		if bad.Code < 400 || bad.Header().Get("Cache-Control") != "no-store" || strings.Contains(bad.Body.String(), realmEnvelope.Data.Token) {
			t.Fatal("invalid realm request succeeded or leaked token")
		}
	}
	send("DELETE", "/ceph/user", map[string]any{"entity": "client.backup"})
	for _, spec := range runner.specs {
		if strings.Contains(fmt.Sprint(spec.Args), "sensitive-import-key") {
			t.Fatal("import key in argv")
		}
	}
	audits, err := db.ListAuditEvents(context.Background(), cluster.ID, 100)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, audit := range audits {
		encoded, _ := json.Marshal(audit)
		if strings.Contains(string(encoded), realmEnvelope.Data.Token) || strings.Contains(string(encoded), "realm-secret") {
			t.Fatal("realm token persisted in audit")
		}
		if audit.ParametersJSON != nil && (strings.Contains(*audit.ParametersJSON, "sensitive-import-key") || strings.Contains(*audit.ParametersJSON, "sensitive-config-value")) {
			t.Fatal("import secret entered audit")
		}
		if audit.Action == "ceph_user.import" {
			found = true
		}
	}
	if !found {
		t.Fatal("import action not audited")
	}
	h.AuthEnabled = func() bool { return true }
	rec = httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest("GET", "/api/v1/ceph/users/export", strings.NewReader(`{"cluster_id":1,"entities":["client.backup"]}`)))
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("unauthenticated export: %d", rec.Code)
	}
}
