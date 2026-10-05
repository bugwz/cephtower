package external

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	endpointservice "cephtower/backend/internal/service/endpoint"
)

func TestRGWPerfSnapshot(t *testing.T) {
	s, endpoints, cluster := externalTestService(t)
	ctx := context.Background()
	if _, err := s.RGWPerf(ctx, cluster.ID, "1"); err == nil {
		t.Fatal("missing endpoint accepted")
	}
	if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "prometheus", URL: "https://prometheus.example.test"}); err != nil {
		t.Fatal(err)
	}
	calls := 0
	body := `{"status":"success","data":{"resultType":"vector","result":[{"metric":{"__name__":"ceph_rgw_req","instance_id":"1","job":"mgr"},"value":[1,"9007199254740993"]}]}}`
	s.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Query().Get("query") != `{__name__=~"ceph_rgw_.*",instance_id="1"}` {
			t.Fatal("unscoped query")
		}
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
	})
	result, err := s.RGWPerf(ctx, cluster.ID, "1")
	encoded, _ := json.Marshal(result)
	if err != nil || !strings.Contains(string(encoded), `"available":true`) || !strings.Contains(string(encoded), `"9007199254740993"`) || !strings.Contains(string(encoded), `"source":"prometheus"`) {
		t.Fatalf("invalid snapshot: %s %v", encoded, err)
	}
	for _, id := range []string{"", "bad\n", strings.Repeat("a", 257)} {
		if _, err := s.RGWPerf(ctx, cluster.ID, id); err == nil {
			t.Fatal("invalid id accepted")
		}
	}
	if _, err := s.RGWPerf(ctx, 0, "1"); err == nil || calls != 1 {
		t.Fatal("invalid scope requested")
	}
	body = `{"status":"success","data":{"resultType":"vector","result":[]}}`
	result, err = s.RGWPerf(ctx, cluster.ID, "1")
	encoded, _ = json.Marshal(result)
	if err != nil || !strings.Contains(string(encoded), `"available":false`) || !strings.Contains(string(encoded), `"series":[]`) {
		t.Fatal("empty result became available")
	}
	body = `private-diagnostic`
	if result, err = s.RGWPerf(ctx, cluster.ID, "1"); err == nil || result != nil || strings.Contains(err.Error(), "private-") {
		t.Fatal("failure leaked or hidden")
	}
}
