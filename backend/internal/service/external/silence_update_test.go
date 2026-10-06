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

func TestSilenceUpdatePreflightAndNativePayload(t *testing.T) {
	for _, scenario := range []string{"ok", "pending", "stale", "expired", "missing", "duplicate", "read-error", "write-error", "missing-result-id", "bad-key"} {
		t.Run(scenario, func(t *testing.T) {
			s, endpoints, cluster := externalTestService(t)
			ctx := context.Background()
			if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "alertmanager", URL: "https://alertmanager.example.test"}); err != nil {
				t.Fatal(err)
			}
			reads, writes := 0, 0
			s.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				body, status := `{"silenceID":"updated-id"}`, 200
				if r.Method == http.MethodGet {
					reads++
					if r.URL.Path != "/api/v2/silences" {
						t.Fatalf("path %s", r.URL)
					}
					state, updated := "active", "2026-01-01T00:00:00Z"
					if scenario == "pending" {
						state = "pending"
					}
					if scenario == "expired" {
						state = "expired"
					}
					if scenario == "stale" {
						updated = "2026-01-01T00:01:00Z"
					}
					row := `{"id":"original","status":{"state":"` + state + `"},"updatedAt":"` + updated + `"}`
					body = "[" + row + "]"
					if scenario == "missing" {
						body = "[]"
					}
					if scenario == "duplicate" {
						body = "[" + row + "," + row + "]"
					}
					if scenario == "read-error" {
						status = 503
					}
				} else {
					writes++
					if r.Method != http.MethodPost || r.URL.Path != "/api/v2/silences" {
						t.Fatalf("unexpected write %s %s", r.Method, r.URL)
					}
					var payload map[string]any
					if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
						t.Fatal(err)
					}
					if payload["id"] != "original" || payload["comment"] != "changed" {
						t.Fatalf("payload %#v", payload)
					}
					for _, field := range []string{"expected_updated_at", "silence_id", "cluster_id", "updatedAt", "status"} {
						if _, ok := payload[field]; ok {
							t.Fatalf("leaked %s", field)
						}
					}
					match := payload["matchers"].([]any)[0].(map[string]any)
					if match["isRegex"] != true || match["isEqual"] != false {
						t.Fatalf("matcher %#v", match)
					}
					if scenario == "write-error" {
						status = 503
					}
					if scenario == "missing-result-id" {
						body = "{}"
					}
				}
				return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body)), Request: r}, nil
			})
			key := "silence/original"
			if scenario == "bad-key" {
				key = "silence/parent/original"
			}
			result, err := s.Execute(ctx, Request{ClusterID: cluster.ID, Action: "silence.update", ResourceKey: key, Parameters: map[string]any{
				"expected_updated_at": "2026-01-01T00:00:00Z", "id": "untrusted-id", "startsAt": "2026-01-01T00:00:00Z", "endsAt": "2026-01-01T02:00:00Z", "createdBy": "operator", "comment": "changed",
				"matchers": []any{map[string]any{"name": "instance", "value": "host.*", "isRegex": true, "isEqual": false}},
			}})
			if scenario == "ok" || scenario == "pending" {
				if err != nil || writes != 1 || result.Details.(map[string]any)["silence_id"] != "updated-id" {
					t.Fatalf("result %#v err %v writes %d", result, err, writes)
				}
			} else {
				if err == nil {
					t.Fatal("expected failure")
				}
				wantWrites := 0
				if scenario == "write-error" || scenario == "missing-result-id" {
					wantWrites = 1
				}
				if writes != wantWrites {
					t.Fatalf("unexpected writes %d", writes)
				}
			}
			if scenario == "bad-key" && reads != 0 {
				t.Fatal("invalid key triggered read")
			}
		})
	}
}
