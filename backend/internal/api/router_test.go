package api

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"cephtower/backend/internal/integrations/ceph"
)

func TestCephV1Routes(t *testing.T) {
	client := &fakeCephClient{
		rawPayload: json.RawMessage(`{"ok":true}`),
		hosts: []ceph.Host{
			{Hostname: "node-a", Addr: "10.0.0.1"},
		},
	}
	mux := http.NewServeMux()
	server := &Server{ceph: client}
	server.registerAPIRouter(mux)

	t.Run("healthz is under api v1", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/healthz", nil)
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d: %s", rr.Code, http.StatusOK, rr.Body.String())
		}
	})

	t.Run("list hosts maps query options", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/host?limit=5&offset=2&facts=true&include_service_instances=true&search=node&sort=hostname", nil)
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d: %s", rr.Code, http.StatusOK, rr.Body.String())
		}
		if client.listHostsOptions.Limit == nil || *client.listHostsOptions.Limit != 5 {
			t.Fatalf("limit = %#v, want 5", client.listHostsOptions.Limit)
		}
		if client.listHostsOptions.Offset == nil || *client.listHostsOptions.Offset != 2 {
			t.Fatalf("offset = %#v, want 2", client.listHostsOptions.Offset)
		}
		if client.listHostsOptions.Facts == nil || !*client.listHostsOptions.Facts {
			t.Fatalf("facts = %#v, want true", client.listHostsOptions.Facts)
		}
		if client.listHostsOptions.IncludeServiceInstances == nil || !*client.listHostsOptions.IncludeServiceInstances {
			t.Fatalf("include_service_instances = %#v, want true", client.listHostsOptions.IncludeServiceInstances)
		}
		if client.listHostsOptions.Search != "node" || client.listHostsOptions.Sort != "hostname" {
			t.Fatalf("search/sort = %q/%q, want node/hostname", client.listHostsOptions.Search, client.listHostsOptions.Sort)
		}
	})

	t.Run("proxy maps project pool route", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/api/v1/pool/rbd?stats=true", strings.NewReader(`{"application_metadata":"rbd"}`))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		mux.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d: %s", rr.Code, http.StatusOK, rr.Body.String())
		}
		if client.rawMethod != http.MethodPut {
			t.Fatalf("raw method = %q, want %q", client.rawMethod, http.MethodPut)
		}
		if client.rawPath != "/api/pool/rbd" {
			t.Fatalf("raw path = %q, want /api/pool/rbd", client.rawPath)
		}
		if client.rawQuery.Get("stats") != "true" {
			t.Fatalf("stats query = %q, want true", client.rawQuery.Get("stats"))
		}
		body, ok := client.rawBody.(json.RawMessage)
		if !ok {
			t.Fatalf("raw body type = %T, want json.RawMessage", client.rawBody)
		}
		if string(body) != `{"application_metadata":"rbd"}` {
			t.Fatalf("raw body = %s", body)
		}
	})
}

type fakeCephClient struct {
	rawMethod string
	rawPath   string
	rawQuery  url.Values
	rawBody   any

	rawPayload       json.RawMessage
	hosts            []ceph.Host
	listHostsOptions ceph.ListHostsOptions
	err              error
}

func (c *fakeCephClient) ClusterSummary(context.Context) (ceph.ClusterSummary, error) {
	return ceph.ClusterSummary{HealthStatus: "HEALTH_OK"}, c.err
}

func (c *fakeCephClient) Version(context.Context) (string, error) {
	return "ceph version test", c.err
}

func (c *fakeCephClient) HealthFull(context.Context) (map[string]any, error) {
	return map[string]any{"status": "full"}, c.err
}

func (c *fakeCephClient) HealthMinimal(context.Context) (map[string]any, error) {
	return map[string]any{"status": "minimal"}, c.err
}

func (c *fakeCephClient) ListHosts(_ context.Context, options ceph.ListHostsOptions) ([]ceph.Host, error) {
	c.listHostsOptions = options
	return c.hosts, c.err
}

func (c *fakeCephClient) HostDetails(context.Context, string) (map[string]any, error) {
	return map[string]any{}, c.err
}

func (c *fakeCephClient) CreateHost(context.Context, ceph.HostRequest) error {
	return c.err
}

func (c *fakeCephClient) UpdateHost(context.Context, string, ceph.UpdateHostRequest) error {
	return c.err
}

func (c *fakeCephClient) DeleteHost(context.Context, string) error {
	return c.err
}

func (c *fakeCephClient) HostDaemons(context.Context, string) ([]map[string]any, error) {
	return []map[string]any{}, c.err
}

func (c *fakeCephClient) HostDevices(context.Context, string) ([]map[string]any, error) {
	return []map[string]any{}, c.err
}

func (c *fakeCephClient) HostInventory(context.Context, string) (map[string]any, error) {
	return map[string]any{}, c.err
}

func (c *fakeCephClient) ListOSDs(context.Context, ceph.ListOSDsOptions) ([]map[string]any, error) {
	return []map[string]any{}, c.err
}

func (c *fakeCephClient) GetOSD(context.Context, string) (map[string]any, error) {
	return map[string]any{}, c.err
}

func (c *fakeCephClient) OSDFlags(context.Context) ([]string, error) {
	return []string{"noout"}, c.err
}

func (c *fakeCephClient) ListDaemons(context.Context, string) ([]map[string]any, error) {
	return []map[string]any{}, c.err
}

func (c *fakeCephClient) ApplyDaemonAction(context.Context, string, ceph.DaemonActionRequest) error {
	return c.err
}

func (c *fakeCephClient) Raw(_ context.Context, method string, path string, query url.Values, body any) (json.RawMessage, error) {
	c.rawMethod = method
	c.rawPath = path
	c.rawQuery = query
	c.rawBody = body
	if c.err != nil {
		return nil, c.err
	}
	if len(c.rawPayload) == 0 {
		return json.RawMessage(`{}`), nil
	}
	return c.rawPayload, nil
}
