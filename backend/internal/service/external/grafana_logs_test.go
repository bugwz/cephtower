package external

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"testing"

	endpointservice "cephtower/backend/internal/service/endpoint"
)

func TestGrafanaLogsURL(t *testing.T) {
	for _, base := range []string{"https://grafana.test", "https://grafana.test/prefix/", "https://grafana.test/prefix?token=secret#fragment"} {
		u, err := url.Parse(grafanaLogsURL(base))
		if err != nil || !strings.HasSuffix(u.Path, "/explore") || u.Fragment != "" || u.Query().Get("token") != "" || u.Query().Get("orgId") != "1" {
			t.Fatalf("invalid explore URL: %v %v", u, err)
		}
		if strings.Contains(base, "/prefix") && u.Path != "/prefix/explore" {
			t.Fatal("lost reverse proxy prefix")
		}
		var state map[string]any
		if json.Unmarshal([]byte(u.Query().Get("left")), &state) != nil || state["datasource"] != "Loki" {
			t.Fatal("invalid Explore state")
		}
		if query := state["queries"].([]any)[0].(map[string]any); len(query) != 1 || query["refId"] != "A" {
			t.Fatal("must not execute a cross-cluster query")
		}
	}
	for _, base := range []string{"javascript:alert(1)", "/relative", "https://user:secret@grafana.test", "ftp://grafana.test", "%"} {
		if grafanaLogsURL(base) != "" {
			t.Fatalf("unsafe URL accepted: %s", base)
		}
	}
}

func TestGrafanaReadIncludesLogsEntry(t *testing.T) {
	s, endpoints, cluster := externalTestService(t)
	_, err := endpoints.CreateEndpoint(context.Background(), cluster.ID, endpointservice.EndpointInput{Kind: "grafana", URL: "https://grafana.test/prefix"})
	if err != nil {
		t.Fatal(err)
	}
	s.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(`[]`))}, nil
	})
	result, err := s.Read(context.Background(), cluster.ID, "grafana", "", nil)
	if err != nil {
		t.Fatal(err)
	}
	meta := result.(map[string]any)["meta"].(map[string]any)
	if meta["smb_overview_url"] != "https://grafana.test/prefix/d/feem6ehrmi2o0b/smb-overview" {
		t.Fatal("missing SMB dashboard entry")
	}
	if meta["logs_explore_url"] != grafanaLogsURL("https://grafana.test/prefix") {
		t.Fatalf("missing selected cluster entry: %v", meta)
	}
}

func TestGrafanaSMBURL(t *testing.T) {
	if got := grafanaSMBURL("https://grafana.test/prefix/?secret=value#fragment"); got != "https://grafana.test/prefix/d/feem6ehrmi2o0b/smb-overview" {
		t.Fatal(got)
	}
	for _, base := range []string{"/relative", "javascript:alert(1)", "https://user:secret@grafana.test", "ftp://grafana.test", "%"} {
		if grafanaSMBURL(base) != "" {
			t.Fatal("unsafe URL accepted")
		}
	}
}
