package monitoring

import (
	"context"
	"net/http"
	"testing"
)

func TestNativeAlertGroups(t *testing.T) {
	for _, body := range []string{`[]`, `[{"labels":{"cluster":"ceph"},"receiver":{"name":"email"},"alerts":[{"fingerprint":"a","status":{"state":"active"}},{"fingerprint":"b","status":{"state":"suppressed"}}]}]`, `null`, `[null]`, `[{}]`, `[{"alerts":[null]}]`} {
		t.Run(body, func(t *testing.T) {
			client, err := New("https://alerts.example.test/proxy", "token", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path != "/proxy/api/v2/alerts/groups" || r.Method != "GET" || r.Header.Get("Authorization") != "Bearer token" {
					t.Fatalf("request %s", r.URL)
				}
				return jsonResponse(200, body), nil
			})})
			if err != nil {
				t.Fatal(err)
			}
			groups, err := client.AlertGroups(context.Background())
			if body == `[]` {
				if err != nil || groups == nil || len(groups) != 0 {
					t.Fatalf("empty: %v", err)
				}
				return
			}
			if len(body) > 100 {
				if err != nil || len(groups) != 1 || groups[0].Receiver.Name != "email" || groups[0].Labels["cluster"] != "ceph" || len(groups[0].Alerts) != 2 || groups[0].Alerts[1].Status.State != "suppressed" {
					t.Fatalf("groups %#v err %v", groups, err)
				}
			} else if err == nil {
				t.Fatal("invalid groups accepted")
			}
		})
	}
}
