package external

import (
	endpointservice "cephtower/backend/internal/service/endpoint"
	"context"
	"net/http"
	"net/url"
	"testing"
)

func TestAlertReadsRequireStoredFSIDBeforeHTTP(t *testing.T) {
	for _, kind := range []string{"alert", "alert_group"} {
		t.Run(kind, func(t *testing.T) {
			s, endpoints, cluster := externalTestServiceWithFSID(t, nil)
			ctx := context.Background()
			if _, err := endpoints.CreateEndpoint(ctx, cluster.ID, endpointservice.EndpointInput{Kind: "alertmanager", URL: "https://alertmanager.example.test"}); err != nil {
				t.Fatal(err)
			}
			s.transport = externalRoundTripFunc(func(r *http.Request) (*http.Response, error) {
				t.Fatal("unscoped HTTP request issued")
				return nil, nil
			})
			if _, err := s.Read(ctx, cluster.ID, kind, "", url.Values{"filter": []string{"cluster=other"}, "fsid": []string{"other"}}); err == nil {
				t.Fatal("caller input substituted for discovered FSID")
			}
		})
	}
}
