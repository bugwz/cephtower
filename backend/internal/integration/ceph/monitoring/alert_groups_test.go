package monitoring

import (
	"context"
	"net/http"
	"strconv"
	"testing"
)

func TestAlertGroupFilterCannotBeBroadened(t *testing.T) {
	for _, fsid := range []string{"", " ", `one",cluster=~".*`, "cluster&filter=other"} {
		calls := 0
		client, err := New("https://alerts.example.test", "", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			filters := r.URL.Query()["filter"]
			if len(filters) != 1 || filters[0] != "cluster="+strconv.Quote(fsid) {
				t.Fatalf("filter injection: %s", r.URL)
			}
			return jsonResponse(200, `[]`), nil
		})})
		if err != nil {
			t.Fatal(err)
		}
		_, err = client.AlertGroups(context.Background(), fsid)
		if fsid == "" || fsid == " " {
			if err == nil || calls != 0 {
				t.Fatal("missing FSID was queried")
			}
		} else if err != nil || calls != 1 {
			t.Fatalf("filter rejected: %v", err)
		}
	}
}

func TestNativeAlertGroups(t *testing.T) {
	for _, body := range []string{`[]`, `[{"labels":{"cluster":"ceph"},"receiver":{"name":"email"},"alerts":[{"fingerprint":"a","status":{"state":"active"}},{"fingerprint":"b","status":{"state":"suppressed"}}]}]`, `null`, `[null]`, `[{}]`, `[{"alerts":[null]}]`} {
		t.Run(body, func(t *testing.T) {
			client, err := New("https://alerts.example.test/proxy", "token", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				if r.URL.Path != "/proxy/api/v2/alerts/groups" || r.URL.Query().Get("filter") != `cluster="test-fsid"` || r.Method != "GET" || r.Header.Get("Authorization") != "Bearer token" {
					t.Fatalf("request %s", r.URL)
				}
				return jsonResponse(200, body), nil
			})})
			if err != nil {
				t.Fatal(err)
			}
			groups, err := client.AlertGroups(context.Background(), "test-fsid")
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
