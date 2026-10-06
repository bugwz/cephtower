package monitoring

import (
	"context"
	"net/http"
	"reflect"
	"testing"
)

func TestGrafanaNativeSearchDashboard(t *testing.T) {
	client, err := New("https://grafana.example.test", "token", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method != "GET" || r.URL.Path != "/api/search" || r.URL.Query().Get("type") != "dash-db" || r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected request %s", r.URL)
		}
		return jsonResponse(200, `[{"id":42,"uid":"ceph-overview","title":"Ceph Overview","uri":"db/ceph-overview","url":"/d/ceph-overview/overview","type":"dash-db","tags":["ceph","cluster"],"folderUid":"ceph","folderTitle":"Ceph","folderUrl":"/dashboards/f/ceph","isStarred":true}]`), nil
	})})
	if err != nil {
		t.Fatal(err)
	}
	rows, err := client.Dashboards(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 1 || rows[0].ID != 42 || rows[0].UID != "ceph-overview" || rows[0].URI != "db/ceph-overview" || rows[0].FolderTitle != "Ceph" || rows[0].FolderUID != "ceph" || rows[0].FolderURL != "/dashboards/f/ceph" || rows[0].Type != "dash-db" || !rows[0].IsStarred || !reflect.DeepEqual(rows[0].Tags, []string{"ceph", "cluster"}) {
		t.Fatalf("lost dashboard fields: %#v", rows)
	}
}

func TestGrafanaRejectsAbsentOrInvalidDashboardLists(t *testing.T) {
	for _, body := range []string{"", "null", "{}", "[null]", "[{}]", `[{"id":"42","uid":"ceph"}]`, "[]"} {
		t.Run(body, func(t *testing.T) {
			client, err := New("https://grafana.example.test", "", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) { return jsonResponse(200, body), nil })})
			if err != nil {
				t.Fatal(err)
			}
			rows, err := client.Dashboards(context.Background())
			if body == "[]" {
				if err != nil || rows == nil || len(rows) != 0 {
					t.Fatalf("empty list: %#v %v", rows, err)
				}
				return
			}
			if err == nil {
				t.Fatalf("accepted invalid response %q", body)
			}
		})
	}
}
