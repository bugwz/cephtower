package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"testing"
)

func TestGrafanaDashboardPagination(t *testing.T) {
	for _, scenario := range []string{"tail", "empty-tail", "overlap", "failure", "null", "cancel"} {
		t.Run(scenario, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			calls := 0
			first := make([]Dashboard, 1000)
			for i := range first {
				first[i] = Dashboard{ID: int64(i + 1), UID: fmt.Sprintf("dashboard-%d", i)}
			}
			encoded, err := json.Marshal(first)
			if err != nil {
				t.Fatal(err)
			}
			client, err := New("https://grafana.example.test", "", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if calls > 2 || r.URL.Query().Get("page") != fmt.Sprint(calls) || r.URL.Query().Get("limit") != "1000" || r.URL.Query().Get("type") != "dash-db" {
					t.Fatalf("unexpected page request %s", r.URL)
				}
				if calls == 1 {
					if scenario == "cancel" {
						cancel()
					}
					return jsonResponse(200, string(encoded)), nil
				}
				switch scenario {
				case "empty-tail":
					return jsonResponse(200, "[]"), nil
				case "overlap":
					return jsonResponse(200, `[{"uid":"dashboard-0"}]`), nil
				case "failure":
					return jsonResponse(503, `{}`), nil
				case "null":
					return jsonResponse(200, `null`), nil
				default:
					return jsonResponse(200, `[{"id":1001,"uid":"last"}]`), nil
				}
			})})
			if err != nil {
				t.Fatal(err)
			}
			rows, err := client.Dashboards(ctx)
			if scenario == "tail" || scenario == "empty-tail" {
				want := 1000
				if scenario == "tail" {
					want++
				}
				if err != nil || len(rows) != want || calls != 2 || rows[999].UID != "dashboard-999" {
					t.Fatalf("rows %d calls %d err %v", len(rows), calls, err)
				}
			} else if err == nil || rows != nil {
				t.Fatalf("partial result accepted: %d rows err %v", len(rows), err)
			}
			if scenario == "cancel" && calls != 1 {
				t.Fatal("request continued after cancellation")
			}
		})
	}
}

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

func TestGrafanaPaginationBounds(t *testing.T) {
	for _, oversized := range []bool{false, true} {
		calls := 0
		client, err := New("https://grafana.example.test", "", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			if calls > 100 {
				t.Fatal("pagination did not stop at safety bound")
			}
			size := 1000
			if oversized {
				size++
			}
			rows := make([]Dashboard, size)
			for i := range rows {
				rows[i].UID = fmt.Sprintf("%d-%d", calls, i)
			}
			body, err := json.Marshal(rows)
			if err != nil {
				t.Fatal(err)
			}
			return jsonResponse(200, string(body)), nil
		})})
		if err != nil {
			t.Fatal(err)
		}
		rows, err := client.Dashboards(context.Background())
		want := 100
		if oversized {
			want = 1
		}
		if err == nil || rows != nil || calls != want {
			t.Fatalf("bound result: rows %d calls %d err %v", len(rows), calls, err)
		}
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
