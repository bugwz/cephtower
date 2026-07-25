package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	ceph "cephtower/backend/internal/service/cephproxy"
)

func TestListHostsMapsQueryOptions(t *testing.T) {
	client := &cephManagementFakeClient{
		hosts: []ceph.Host{{Hostname: "node-a", Addr: "10.0.0.1"}},
	}
	h := New(client, Dependencies{})
	request := httptest.NewRequest(
		http.MethodGet,
		"/api/v1/host?limit=5&offset=2&facts=true&include_service_instances=true&search=node&sort=hostname",
		nil,
	)
	recorder := httptest.NewRecorder()

	h.ListHosts(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	options := client.listHostsOptions
	if options.Limit == nil || *options.Limit != 5 {
		t.Fatalf("limit = %#v, want 5", options.Limit)
	}
	if options.Offset == nil || *options.Offset != 2 {
		t.Fatalf("offset = %#v, want 2", options.Offset)
	}
	if options.Facts == nil || !*options.Facts {
		t.Fatalf("facts = %#v, want true", options.Facts)
	}
	if options.IncludeServiceInstances == nil || !*options.IncludeServiceInstances {
		t.Fatalf("include_service_instances = %#v, want true", options.IncludeServiceInstances)
	}
	if options.Search != "node" || options.Sort != "hostname" {
		t.Fatalf("search/sort = %q/%q, want node/hostname", options.Search, options.Sort)
	}
}

func TestProxyPathMapsProjectRoute(t *testing.T) {
	client := &cephManagementFakeClient{}
	h := New(client, Dependencies{})
	request := httptest.NewRequest(
		http.MethodPut,
		"/api/v1/pool/rbd?stats=true",
		strings.NewReader(`{"application_metadata":"rbd"}`),
	)
	request.Pattern = "PUT /api/v1/pool/{name}"
	request.SetPathValue("name", "rbd")
	recorder := httptest.NewRecorder()

	h.ProxyPath(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if client.rawMethod != http.MethodPut || client.rawPath != "/api/pool/rbd" {
		t.Fatalf("raw request = %s %s, want PUT /api/pool/rbd", client.rawMethod, client.rawPath)
	}
	if client.rawQuery.Get("stats") != "true" {
		t.Fatalf("stats query = %q, want true", client.rawQuery.Get("stats"))
	}
	body, ok := client.rawBody.(json.RawMessage)
	if !ok || string(body) != `{"application_metadata":"rbd"}` {
		t.Fatalf("raw body = %T(%s), want json.RawMessage", client.rawBody, body)
	}
}
