package ceph

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"cephtower/backend/internal/config"
	"cephtower/backend/internal/integrations/ceph/typed"
	"gopkg.in/yaml.v3"
)

func TestClusterSummaryUsesDashboardSummaryWithToken(t *testing.T) {
	var authCalls int
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/api/auth":
			authCalls++
			if r.Method != http.MethodPost {
				t.Fatalf("auth method = %s, want POST", r.Method)
			}
			return testJSONResponse(http.StatusCreated, map[string]any{
				"token":             "test-token",
				"username":          "admin",
				"permissions":       map[string][]string{"cephfs": {}},
				"pwdExpirationDate": "",
				"pwdUpdateRequired": false,
				"sso":               false,
			}), nil
		case "/api/summary":
			if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
				t.Fatalf("Authorization = %q, want bearer token", got)
			}
			return testJSONResponse(http.StatusOK, map[string]any{
				"health_status":       "HEALTH_OK",
				"version":             "20.2.2",
				"mgr_id":              "a",
				"mgr_host":            "node-1",
				"have_mon_connection": "true",
				"executing_tasks":     []string{},
				"finished_tasks":      []any{},
				"rbd_mirroring":       map[string]int{"warnings": 0, "errors": 0},
			}), nil
		default:
			return testStringResponse(http.StatusNotFound, "not found"), nil
		}
	})

	client := NewDashboardClientWithHTTPClient(config.CephDashboardConfig{
		BaseURL:  "https://ceph.example.com",
		Username: "admin",
		Password: "password",
	}, &http.Client{Transport: transport})

	summary, err := client.ClusterSummary(context.Background())
	if err != nil {
		t.Fatalf("ClusterSummary() returned error: %v", err)
	}

	if summary.HealthStatus != "HEALTH_OK" || summary.Version != "20.2.2" {
		t.Fatalf("summary = %#v, want health and version from /api/summary", summary)
	}
	if authCalls != 1 {
		t.Fatalf("auth calls = %d, want 1", authCalls)
	}
}

func TestDoRefreshesTokenAfterUnauthorized(t *testing.T) {
	var authCalls int
	var summaryCalls int
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/api/auth":
			authCalls++
			return testJSONResponse(http.StatusCreated, map[string]any{
				"token":             "token-" + string(rune('0'+authCalls)),
				"username":          "admin",
				"permissions":       map[string][]string{"cephfs": {}},
				"pwdExpirationDate": "",
				"pwdUpdateRequired": false,
				"sso":               false,
			}), nil
		case "/api/summary":
			summaryCalls++
			if summaryCalls == 1 {
				return testStringResponse(http.StatusUnauthorized, "expired"), nil
			}
			if got := r.Header.Get("Authorization"); got != "Bearer token-2" {
				t.Fatalf("Authorization after retry = %q, want refreshed token", got)
			}
			return testJSONResponse(http.StatusOK, map[string]any{
				"health_status": "HEALTH_WARN",
				"version":       "20.2.2",
			}), nil
		default:
			return testStringResponse(http.StatusNotFound, "not found"), nil
		}
	})

	client := NewDashboardClientWithHTTPClient(config.CephDashboardConfig{
		BaseURL:  "https://ceph.example.com",
		Username: "admin",
		Password: "password",
	}, &http.Client{Transport: transport})

	if _, err := client.ClusterSummary(context.Background()); err != nil {
		t.Fatalf("ClusterSummary() returned error: %v", err)
	}
	if authCalls != 2 {
		t.Fatalf("auth calls = %d, want refresh after 401", authCalls)
	}
}

func TestGeneratedOperationRendersPathAndQuery(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.EscapedPath() != "/api/host/node%2F1/devices" {
			t.Fatalf("path = %s, want escaped path parameter", r.URL.EscapedPath())
		}
		if got := r.URL.Query().Get("kind"); got != "ssd" {
			t.Fatalf("query kind = %q, want ssd", got)
		}
		return testJSONResponse(http.StatusOK, []map[string]string{{"id": "dev-1"}}), nil
	})

	client := NewDashboardClientWithHTTPClient(
		config.CephDashboardConfig{BaseURL: "https://ceph.example.com"},
		&http.Client{Transport: transport},
	)
	payload, err := client.API().GetHostByHostnameDevices(context.Background(), OperationRequest{
		Path:  map[string]string{"hostname": "node/1"},
		Query: url.Values{"kind": []string{"ssd"}},
	})
	if err != nil {
		t.Fatalf("GetHostByHostnameDevices() returned error: %v", err)
	}
	if !strings.Contains(string(payload), "dev-1") {
		t.Fatalf("payload = %s, want raw JSON response", payload)
	}
}

func TestTypedOperationUsesTypedRequestAndResponse(t *testing.T) {
	transport := roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != http.MethodPut {
			t.Fatalf("method = %s, want PUT", r.Method)
		}
		if r.URL.EscapedPath() != "/api/host/node%2F1" {
			t.Fatalf("path = %s, want escaped typed path parameter", r.URL.EscapedPath())
		}

		data, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}
		if !strings.Contains(string(data), `"maintenance":true`) {
			t.Fatalf("request body = %s, want typed body JSON", data)
		}

		return testJSONResponse(http.StatusOK, map[string]string{"status": "ok"}), nil
	})

	client := NewDashboardClientWithHTTPClient(
		config.CephDashboardConfig{BaseURL: "https://ceph.example.com"},
		&http.Client{Transport: transport},
	)
	response, err := client.TypedAPI().PutHostByHostname(context.Background(), typed.PutHostByHostnameRequest{
		Hostname: "node/1",
		Body: typed.PutHostByHostnameBody{
			Maintenance: true,
		},
	})
	if err != nil {
		t.Fatalf("TypedAPI().PutHostByHostname() returned error: %v", err)
	}
	if string(response["status"]) != `"ok"` {
		t.Fatalf("typed response = %#v, want decoded JSON object", response)
	}
}

func TestGeneratedOperationRequiresPathParams(t *testing.T) {
	client := NewDashboardClient(config.CephDashboardConfig{BaseURL: "https://ceph.example.com"})
	_, err := client.API().GetHostByHostnameDevices(context.Background(), OperationRequest{})
	if err == nil {
		t.Fatal("GetHostByHostnameDevices() error = nil, want missing path parameter error")
	}
	if !strings.Contains(err.Error(), "hostname") {
		t.Fatalf("error = %v, want hostname mentioned", err)
	}
}

func TestAPIErrorIncludesStatusAndBody(t *testing.T) {
	client := NewDashboardClientWithHTTPClient(
		config.CephDashboardConfig{BaseURL: "https://ceph.example.com"},
		&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
			return testStringResponse(http.StatusForbidden, "permission denied"), nil
		})},
	)
	err := client.Do(context.Background(), Request{
		Method: http.MethodGet,
		Path:   "/api/summary",
	}, nil)

	var apiErr *APIError
	if !errors.As(err, &apiErr) {
		t.Fatalf("error = %T, want *APIError", err)
	}
	if apiErr.StatusCode != http.StatusForbidden || !strings.Contains(apiErr.Body, "permission denied") {
		t.Fatalf("APIError = %#v, want status and response body", apiErr)
	}
}

func TestGeneratedClientCoversOpenAPIOperations(t *testing.T) {
	openapiPath := filepath.Clean("../../../../docs/references/ceph/src/pybind/mgr/dashboard/openapi.yaml")
	data, err := os.ReadFile(openapiPath)
	if err != nil {
		t.Fatalf("read openapi fixture: %v", err)
	}

	var openapi struct {
		Paths map[string]map[string]any `yaml:"paths"`
	}
	if err := yaml.Unmarshal(data, &openapi); err != nil {
		t.Fatalf("parse openapi fixture: %v", err)
	}

	want := 0
	for _, methods := range openapi.Paths {
		for method := range methods {
			switch strings.ToLower(method) {
			case "get", "post", "put", "patch", "delete":
				want++
			}
		}
	}

	generatedFiles, err := generatedEndpointFiles()
	if err != nil {
		t.Fatalf("glob generated files: %v", err)
	}

	methodRE := regexp.MustCompile(`func \(c \*Client\) [A-Z][A-Za-z0-9]*\(ctx context\.Context, request OperationRequest\) \(json\.RawMessage, error\)`)
	got := 0
	for _, path := range generatedFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read generated file %s: %v", path, err)
		}
		got += len(methodRE.FindAll(data, -1))
	}

	if got != want {
		t.Fatalf("generated operation methods = %d, want %d from OpenAPI", got, want)
	}

	typedFiles, err := typedEndpointFiles()
	if err != nil {
		t.Fatalf("list typed endpoint files: %v", err)
	}

	typedMethodRE := regexp.MustCompile(`func \(c \*Client\) [A-Z][A-Za-z0-9]*\(ctx context\.Context, request [A-Z][A-Za-z0-9]*Request\) \([A-Z][A-Za-z0-9]*Response, error\)`)
	got = 0
	for _, path := range typedFiles {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read typed file %s: %v", path, err)
		}
		got += len(typedMethodRE.FindAll(data, -1))
	}

	if got != want {
		t.Fatalf("typed operation methods = %d, want %d from OpenAPI", got, want)
	}
}

func generatedEndpointFiles() ([]string, error) {
	return endpointFiles("endpoints")
}

func typedEndpointFiles() ([]string, error) {
	return endpointFiles("typed")
}

func endpointFiles(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, entry := range entries {
		if entry.IsDir() || entry.Name() == "client.go" || !strings.HasSuffix(entry.Name(), ".go") {
			continue
		}
		files = append(files, filepath.Join(dir, entry.Name()))
	}

	return files, nil
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func testJSONResponse(status int, payload any) *http.Response {
	data, _ := json.Marshal(payload)
	return testResponse(status, string(data))
}

func testStringResponse(status int, body string) *http.Response {
	return testResponse(status, body)
}

func testResponse(status int, body string) *http.Response {
	return &http.Response{
		StatusCode: status,
		Status:     http.StatusText(status),
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
