package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"cephtower/backend/internal/api"
	"cephtower/backend/internal/api/v1/handler"
	"cephtower/backend/internal/config"
	authservice "cephtower/backend/internal/service/auth"
	clusterservice "cephtower/backend/internal/service/cluster"
	"cephtower/backend/internal/store"
)

func TestAuthorizationEnforcesRolesAndClusterIsolation(t *testing.T) {
	ctx := context.Background()
	db, err := store.Open(config.DatabaseConfig{EncryptionKey: contractKey, Engine: store.EngineSQLite, SQLite: config.SQLiteConfig{Name: "authorization.db"}}, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close(db)
	auth := authservice.New(func() *store.Database { return db }, func() config.Config {
		return config.Config{Database: config.DatabaseConfig{EncryptionKey: contractKey}}
	})
	if err := auth.EnsureRoles(ctx); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	clusterIDs := []uint64{}
	for _, name := range []string{"allowed", "hidden"} {
		row := store.CephCluster{Name: name, MonitorAddresses: "mon:6789", ClientUsername: "client." + name, ClientKey: "encrypted", CreatedAt: now, UpdatedAt: now}
		if err := db.CreateCluster(ctx, &row); err != nil {
			t.Fatal(err)
		}
		clusterIDs = append(clusterIDs, row.ID)
	}
	// Sessions are seeded directly to isolate authorization from password verification.
	tokens := map[string]string{}
	users := map[string]uint64{}
	for _, role := range []string{"viewer", "operator", "storage-admin", "security-admin", "cluster-admin", "scoped-admin", "unbound"} {
		user := store.User{Username: role, DisplayName: role, Password: "unused", Status: "active", CreatedAt: now, UpdatedAt: now}
		if err := db.CreateUser(ctx, &user); err != nil {
			t.Fatal(err)
		}
		if role != "unbound" {
			var scope *uint64
			bindingRole := role
			if role == "scoped-admin" {
				scope = &clusterIDs[0]
				bindingRole = "cluster-admin"
			}
			if err := db.BindUserRole(ctx, user.ID, bindingRole, scope, nil); err != nil {
				t.Fatal(err)
			}
		}
		token := "token-" + role
		session := store.UserSession{ID: role, UserID: user.ID, TokenHash: store.SHA256(token), ExpiresAt: now.Add(time.Hour), CreatedAt: now, LastSeenAt: now}
		if err := db.CreateSessionAndTouchUser(ctx, &session, now); err != nil {
			t.Fatal(err)
		}
		tokens[role], users[role] = token, user.ID
	}
	clusters := clusterservice.New(func() *store.Database { return db }, contractKey, unusedProvider{})
	defer clusters.Stop()
	h := handler.New(handler.Dependencies{Auth: auth, Clusters: clusters, Database: func() *store.Database { return db }})
	routes := api.NewAPI(h).Routes()
	send := func(server http.Handler, role, method, path, body string, override bool) *httptest.ResponseRecorder {
		request := httptest.NewRequest(method, "/api/v1"+path, strings.NewReader(body))
		request.Header.Set("Authorization", "Bearer "+tokens[role])
		if override {
			request.Header.Set("X-HTTP-Method-Override", "GET")
		}
		response := httptest.NewRecorder()
		server.ServeHTTP(response, request)
		return response
	}
	for _, tc := range []struct {
		role, method, path string
		allowed            bool
	}{
		{"viewer", "POST", "/user", false}, {"viewer", "POST", "/role/binding", false},
		{"viewer", "DELETE", "/pool", false}, {"viewer", "GET", "/credentials", false},
		{"viewer", "GET", "/ceph/users/export", false}, {"viewer", "GET", "/host/ssh", false},
		{"viewer", "GET", "/pools", true}, {"operator", "POST", "/resource/refresh", true},
		{"operator", "DELETE", "/pool", false}, {"storage-admin", "POST", "/filesystem", true},
		{"storage-admin", "POST", "/rgw/user", true}, {"storage-admin", "PUT", "/configuration/value", false},
		{"security-admin", "POST", "/role/binding", true}, {"security-admin", "POST", "/user", true},
		{"security-admin", "DELETE", "/pool", false}, {"cluster-admin", "DELETE", "/pool", true},
		{"scoped-admin", "POST", "/user", false}, {"scoped-admin", "POST", "/cluster", false},
		{"scoped-admin", "POST", "/role/binding", true}, {"unbound", "GET", "/pools", false},
	} {
		t.Run(tc.role+" "+tc.method+" "+tc.path, func(t *testing.T) {
			called := false
			endpoint := h.PrepareRoute(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { called = true; w.WriteHeader(204) }))
			response := send(endpoint, tc.role, tc.method, tc.path, fmt.Sprintf(`{"cluster_id":%d}`, clusterIDs[0]), false)
			if tc.allowed {
				if response.Code != 204 || !called {
					t.Fatalf("authorized request rejected: %s", response.Body.String())
				}
			} else if response.Code != 403 || called {
				t.Fatalf("unauthorized request reached handler: %d %s", response.Code, response.Body.String())
			}
		})
	}
	for _, body := range []string{`{}`, `{"cluster_id":0}`, `{"cluster_id":"1"}`, `{"cluster_id":1.5}`, `{"cluster_id":9007199254740993}`, `{"cluster_id":1,"Cluster_ID":2}`} {
		response := send(routes, "cluster-admin", "GET", "/cluster", body, false)
		if response.Code != 400 {
			t.Fatalf("ambiguous scope accepted: %s: %d", body, response.Code)
		}
	}
	// Real router checks, including browser GET override and filter option leakage.
	for _, suffix := range []string{"", "?filter_options=1&fields=name,client_username"} {
		response := send(routes, "scoped-admin", "GET", "/clusters"+suffix, "", false)
		if response.Code != 200 || strings.Contains(response.Body.String(), "hidden") || !strings.Contains(response.Body.String(), "allowed") {
			t.Fatalf("cluster leakage: %s", response.Body.String())
		}
	}
	response := send(routes, "unbound", "GET", "/clusters", "", false)
	if response.Code != 200 || strings.Contains(response.Body.String(), "allowed") {
		t.Fatalf("unbound cluster list: %s", response.Body.String())
	}
	response = send(routes, "scoped-admin", "POST", "/cluster", fmt.Sprintf(`{"cluster_id":%d}`, clusterIDs[1]), true)
	if response.Code != 403 {
		t.Fatalf("cross-cluster override: %d %s", response.Code, response.Body.String())
	}
	response = send(routes, "viewer", "POST", "/user", `{"username":"intruder","display_name":"Intruder","password":"password-123","role":"cluster-admin"}`, false)
	if response.Code != 403 {
		t.Fatalf("viewer created administrator: %s", response.Body.String())
	}
	if _, err := db.FindUserByUsername(ctx, "intruder"); err == nil {
		t.Fatal("unauthorized user persisted")
	}
	response = send(routes, "cluster-admin", "GET", "/auth/me", "", false)
	var me struct {
		Data struct {
			ID   uint64 `json:"id"`
			Role string `json:"role"`
		} `json:"data"`
	}
	if err := json.Unmarshal(response.Body.Bytes(), &me); err != nil {
		t.Fatal(err)
	}
	if response.Code != 200 || me.Data.ID != users["cluster-admin"] || me.Data.Role != "cluster-admin" {
		t.Fatalf("incorrect current role: %s", response.Body.String())
	}
	bindings, err := db.ListClusterRoleBindings(ctx, clusterIDs[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, binding := range bindings {
		if binding.UserID == users["scoped-admin"] {
			if err := db.DeleteClusterRoleBinding(ctx, clusterIDs[0], binding.ID); err != nil {
				t.Fatal(err)
			}
		}
	}
	response = send(routes, "scoped-admin", "POST", "/cluster", fmt.Sprintf(`{"cluster_id":%d}`, clusterIDs[0]), true)
	if response.Code != 403 {
		t.Fatal("revoked role remained active")
	}
}
