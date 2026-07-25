package handler

// Settings endpoint tests.
import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"cephtower/backend/internal/config"
	cephapi "cephtower/backend/internal/service/cephproxy"
	settingsservice "cephtower/backend/internal/service/settings"
	"cephtower/backend/internal/store"
)

func TestSettingsRedactSensitiveValues(t *testing.T) {
	h, db, admin, fake := newManagementTestAPI(t)
	defer closeTestDB(t, db)

	fake.settings["GRAFANA_API_URL"] = "http://grafana:3000"
	fake.settings["GRAFANA_API_PASSWORD"] = "secret"

	recorder := cephManagementRequest(h.ListSettings, http.MethodGet, "/api/v1/settings", admin, nil, nil)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}

	var settings []settingsservice.Setting
	if err := decodeAPIResponseData(recorder, &settings); err != nil {
		t.Fatalf("decode settings: %v", err)
	}
	if len(settings) != 2 {
		t.Fatalf("settings length = %d, want 2", len(settings))
	}
	byName := map[string]settingsservice.Setting{}
	for _, setting := range settings {
		byName[setting.Name] = setting
	}
	if byName["GRAFANA_API_URL"].Value != "http://grafana:3000" {
		t.Fatalf("non-sensitive value = %#v, want grafana url", byName["GRAFANA_API_URL"].Value)
	}
	password := byName["GRAFANA_API_PASSWORD"]
	if !password.Sensitive || !password.ValueSet || password.Value != "********" {
		t.Fatalf("sensitive setting = %#v, want redacted set password", password)
	}
}

func TestSettingGroupUpdateAuditsRedactedChange(t *testing.T) {
	h, db, admin, fake := newManagementTestAPI(t)
	defer closeTestDB(t, db)

	fake.settings["GRAFANA_API_PASSWORD"] = "old-secret"
	fake.settings["GRAFANA_API_URL"] = "http://grafana:3000"

	recorder := cephManagementRequest(
		h.UpdateSettingGroup,
		http.MethodPut,
		"/api/v1/grafana/settings",
		admin,
		[]byte(`{"GRAFANA_API_PASSWORD":"new-secret"}`),
		nil,
	)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	if fake.settings["GRAFANA_API_PASSWORD"] != "new-secret" {
		t.Fatalf("updated setting = %q, want new-secret", fake.settings["GRAFANA_API_PASSWORD"])
	}

	var change store.CephClusterSettingChange
	if err := db.FindRecord(context.Background(), map[string]any{"setting_name": "GRAFANA_API_PASSWORD"}, &change); err != nil {
		t.Fatalf("load setting change: %v", err)
	}
	if change.Status != "success" || change.OldValueRedacted != `"********"` || change.NewValueRedacted != `"********"` {
		t.Fatalf("change = %#v, want redacted successful audit", change)
	}
	if change.OperatorUserID != admin.ID {
		t.Fatalf("operator = %d, want %d", change.OperatorUserID, admin.ID)
	}
}

func TestSettingGroupRejectsSettingOutsideGroup(t *testing.T) {
	h, db, admin, fake := newManagementTestAPI(t)
	defer closeTestDB(t, db)

	fake.settings["PWD_POLICY_ENABLED"] = true
	recorder := cephManagementRequest(
		h.UpdateSettingGroup,
		http.MethodPut,
		"/api/v1/grafana/settings",
		admin,
		[]byte(`{"PWD_POLICY_ENABLED":false}`),
		nil,
	)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400: %s", recorder.Code, recorder.Body.String())
	}
	envelope, err := decodeAPIResponseEnvelope(recorder)
	if err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Code != http.StatusBadRequest {
		t.Fatalf("envelope code = %d, want 400", envelope.Code)
	}
	if fake.settings["PWD_POLICY_ENABLED"] != true {
		t.Fatalf("setting changed on rejected update")
	}
}

func TestNamedHandlerCallsNativeDashboardAPI(t *testing.T) {
	h, db, user, fake := newManagementTestAPI(t)
	defer closeTestDB(t, db)

	fake.setRaw(http.MethodGet, "/api/nvmeof/gateway/listener_info/nqn.2016-06.io.spdk:cnode1", json.RawMessage(`{"listeners":[]}`))
	recorder := cephManagementRequest(
		h.GetNVMeOFListenerInfo,
		http.MethodGet,
		"/api/v1/nvmeof/gateway/listener-info/nqn.2016-06.io.spdk:cnode1?detail=true",
		user,
		nil,
		map[string]string{"nqn": "nqn.2016-06.io.spdk:cnode1"},
	)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200: %s", recorder.Code, recorder.Body.String())
	}
	if fake.rawMethod != http.MethodGet ||
		fake.rawPath != "/api/nvmeof/gateway/listener_info/nqn.2016-06.io.spdk:cnode1" ||
		fake.rawQuery.Get("detail") != "true" {
		t.Fatalf("raw request = %s %s?%s, want GET listener_info path", fake.rawMethod, fake.rawPath, fake.rawQuery.Encode())
	}
}

func TestDashboardUserHandlerRequiresAdmin(t *testing.T) {
	h, db, _, _ := newManagementTestAPI(t)
	defer closeTestDB(t, db)

	regular := store.User{
		Username:    "operator",
		DisplayName: "Operator",
		Role:        store.UserRoleUser,
		Permissions: `["cluster:read","system:read"]`,
		Enabled:     true,
	}
	recorder := cephManagementRequest(
		h.ListDashboardUsers,
		http.MethodGet,
		"/api/v1/dashboard-user",
		regular,
		nil,
		nil,
	)
	if recorder.Code != http.StatusForbidden {
		t.Fatalf("status = %d, want 403: %s", recorder.Code, recorder.Body.String())
	}
	envelope, err := decodeAPIResponseEnvelope(recorder)
	if err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	if envelope.Code != http.StatusForbidden {
		t.Fatalf("envelope code = %d, want 403", envelope.Code)
	}
}

func newManagementTestAPI(t *testing.T) (*Handler, *store.Database, store.User, *cephManagementFakeClient) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "cephtower.db")
	cfg := config.Config{
		Database: config.DatabaseConfig{
			Engine: store.EngineSQLite,
			SQLite: config.SQLiteConfig{Path: dbPath},
		},
	}
	db, err := store.Open(cfg.Database)
	if err != nil {
		t.Fatalf("Open() returned error: %v", err)
	}

	passwordHash, err := store.HashPassword("ChangeMe123!")
	if err != nil {
		t.Fatalf("HashPassword() returned error: %v", err)
	}
	user := store.User{
		Username:     "admin",
		DisplayName:  "Admin",
		Role:         store.UserRoleAdmin,
		Permissions:  `["cluster:read","storage:read","system:read"]`,
		PasswordHash: passwordHash,
		Enabled:      true,
	}
	if err := db.CreateUser(context.Background(), &user); err != nil {
		t.Fatalf("create user: %v", err)
	}
	cluster := store.CephCluster{
		Name:              "primary",
		MonitorHost:       "10.0.0.1:6789",
		Keyring:           "keyring",
		DashboardUsername: "admin",
		DashboardPassword: "password",
	}
	if err := db.CreateCluster(context.Background(), &cluster); err != nil {
		t.Fatalf("create cluster: %v", err)
	}

	fake := newManagementFakeClient()
	h := New(fake, Dependencies{SettingsService: settingsservice.New(fake, func() *store.Database { return db })})
	return h, db, user, fake
}

func closeTestDB(t *testing.T, db *store.Database) {
	t.Helper()
	if err := store.Close(db); err != nil {
		t.Fatalf("Close() returned error: %v", err)
	}
}

func cephManagementRequest(h http.HandlerFunc, method, path string, user store.User, body []byte, pathValues map[string]string) *httptest.ResponseRecorder {
	recorder := httptest.NewRecorder()
	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(body)
	}
	request := httptest.NewRequest(method, path, reader)
	for key, value := range pathValues {
		request.SetPathValue(key, value)
	}
	request = request.WithContext(ContextWithUser(request.Context(), user))
	h(recorder, request)
	return recorder
}

type cephManagementFakeClient struct {
	rawMethod string
	rawPath   string
	rawQuery  url.Values
	rawBody   any
	hosts     []cephapi.Host

	listHostsOptions cephapi.ListHostsOptions

	rawResponses map[string]json.RawMessage
	settings     map[string]any
}

func newManagementFakeClient() *cephManagementFakeClient {
	return &cephManagementFakeClient{
		rawResponses: map[string]json.RawMessage{},
		settings:     map[string]any{},
	}
}

func (c *cephManagementFakeClient) setRaw(method string, path string, payload json.RawMessage) {
	c.rawResponses[method+" "+path] = payload
}

func (c *cephManagementFakeClient) Raw(_ context.Context, method string, path string, query url.Values, body any) (json.RawMessage, error) {
	c.rawMethod = method
	c.rawPath = path
	c.rawQuery = query
	c.rawBody = body

	if strings.HasPrefix(path, "/api/settings") {
		return c.settingsRaw(method, path, body)
	}
	if payload, ok := c.rawResponses[method+" "+path]; ok {
		return payload, nil
	}
	return json.RawMessage(`{}`), nil
}

func (c *cephManagementFakeClient) settingsRaw(method string, path string, body any) (json.RawMessage, error) {
	switch {
	case method == http.MethodGet && path == "/api/settings":
		items := make([]map[string]any, 0, len(c.settings))
		for name, value := range c.settings {
			items = append(items, map[string]any{
				"name":    name,
				"type":    "str",
				"default": false,
				"value":   value,
			})
		}
		return json.Marshal(items)
	case method == http.MethodGet:
		name := strings.TrimPrefix(path, "/api/settings/")
		return json.Marshal(map[string]any{
			"name":    name,
			"type":    "str",
			"default": false,
			"value":   c.settings[name],
		})
	case method == http.MethodPut && path != "/api/settings":
		name := strings.TrimPrefix(path, "/api/settings/")
		value := valueFromRawBody(body)
		c.settings[name] = value
		return json.Marshal(map[string]any{
			"name":    name,
			"type":    "str",
			"default": false,
			"value":   value,
		})
	default:
		return json.RawMessage(`{}`), nil
	}
}

func valueFromRawBody(body any) any {
	switch typed := body.(type) {
	case map[string]any:
		return typed["value"]
	case json.RawMessage:
		var item map[string]any
		if err := json.Unmarshal(typed, &item); err != nil {
			return nil
		}
		return item["value"]
	default:
		return nil
	}
}

func (c *cephManagementFakeClient) ClusterSummary(context.Context) (cephapi.ClusterSummary, error) {
	return cephapi.ClusterSummary{}, nil
}

func (c *cephManagementFakeClient) Version(context.Context) (string, error) { return "", nil }
func (c *cephManagementFakeClient) HealthFull(context.Context) (map[string]any, error) {
	return map[string]any{}, nil
}

func (c *cephManagementFakeClient) HealthMinimal(context.Context) (map[string]any, error) {
	return map[string]any{}, nil
}

func (c *cephManagementFakeClient) ListHosts(_ context.Context, options cephapi.ListHostsOptions) ([]cephapi.Host, error) {
	c.listHostsOptions = options
	return c.hosts, nil
}

func (c *cephManagementFakeClient) HostDetails(context.Context, string) (map[string]any, error) {
	return map[string]any{}, nil
}

func (c *cephManagementFakeClient) CreateHost(context.Context, cephapi.HostRequest) error {
	return nil
}

func (c *cephManagementFakeClient) UpdateHost(context.Context, string, cephapi.UpdateHostRequest) error {
	return nil
}

func (c *cephManagementFakeClient) DeleteHost(context.Context, string) error { return nil }
func (c *cephManagementFakeClient) HostDaemons(context.Context, string) ([]map[string]any, error) {
	return nil, nil
}

func (c *cephManagementFakeClient) HostDevices(context.Context, string) ([]map[string]any, error) {
	return nil, nil
}

func (c *cephManagementFakeClient) HostInventory(context.Context, string) (map[string]any, error) {
	return map[string]any{}, nil
}

func (c *cephManagementFakeClient) ListOSDs(context.Context, cephapi.ListOSDsOptions) ([]map[string]any, error) {
	return nil, nil
}

func (c *cephManagementFakeClient) GetOSD(context.Context, string) (map[string]any, error) {
	return map[string]any{}, nil
}

func (c *cephManagementFakeClient) OSDFlags(context.Context) ([]string, error) {
	return nil, nil
}

func (c *cephManagementFakeClient) ListDaemons(context.Context, string) ([]map[string]any, error) {
	return nil, nil
}

func (c *cephManagementFakeClient) ApplyDaemonAction(context.Context, string, cephapi.DaemonActionRequest) error {
	return nil
}
