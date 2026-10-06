package external

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	cephdomain "cephtower/backend/internal/domain/ceph"
	endpointservice "cephtower/backend/internal/service/endpoint"
)

func TestSilenceMutationUncertainOutcomeDoesNotAllowRetry(t *testing.T) {
	for _, action := range []string{"silence.create", "silence.delete"} {
		for _, scenario := range []string{"transport", "server", "malformed"} {
			t.Run(action+"/"+scenario, func(t *testing.T) {
				s, endpoints, cluster := externalTestService(t)
				ctx := context.Background()
				if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "alertmanager", URL: "https://alertmanager.example.test"}); err != nil {
					t.Fatal(err)
				}
				writes := 0
				s.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
					writes++
					method, path := http.MethodPost, "/api/v2/silences"
					if action == "silence.delete" {
						method, path = http.MethodDelete, "/api/v2/silence/original"
					}
					if r.Method != method || r.URL.Path != path {
						t.Fatalf("unexpected mutation %s %s", r.Method, r.URL)
					}
					if scenario == "transport" {
						return nil, errors.New("response lost after remote acceptance")
					}
					status := http.StatusOK
					if scenario == "server" {
						status = http.StatusServiceUnavailable
					}
					return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("{")), Request: r}, nil
				})
				_, err := s.Execute(ctx, Request{ClusterID: cluster.ID, Action: action, ResourceKey: "silence/original", Parameters: map[string]any{
					"startsAt": "2026-01-01T00:00:00Z", "endsAt": "2026-01-01T02:00:00Z", "createdBy": "operator", "comment": "maintenance",
					"matchers": []any{map[string]any{"name": "alertname", "value": "CephHealth", "isRegex": false, "isEqual": true}},
				}})
				// DELETE ignores a successful response body; only POST decodes it.
				if action == "silence.delete" && scenario == "malformed" {
					if err != nil || writes != 1 {
						t.Fatalf("successful expiration: %v, writes %d", err, writes)
					}
					return
				}
				var failure *cephdomain.ActionError
				if !errors.As(err, &failure) || failure.Retryable || failure.Code != "alertmanager_failed" || !strings.Contains(failure.Message, "refresh before retrying") || writes != 1 {
					t.Fatalf("unsafe mutation result: %v, writes %d", err, writes)
				}
			})
		}
	}
}
