package monitoring

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestSMBOverviewQueries(t *testing.T) {
	queries := map[string]string{
		"smb_request_duration_rate": "rate(smb_smb2_request_duration_microseconds_sum[5m])",
		"smb_metrics_status": "smb_metrics_status", "smb_sessions": "smb_sessions_total", "smb_users": "smb_users_total", "smb_share_activity": "smb_share_activity",
		"smb_in_bytes_rate": "rate(smb_smb2_request_inbytes[5m])", "smb_out_bytes_rate": "rate(smb_smb2_request_outbytes[5m])", "smb_request_rate": "rate(smb_smb2_request_total[5m])",
	}
	for id, query := range queries {
		for _, history := range []bool{false, true} {
			client, err := New("https://prometheus.test/prefix", "", &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
				path, kind := "/prefix/api/v1/query", "vector"
				if history {
					path, kind = "/prefix/api/v1/query_range", "matrix"
				}
				if r.URL.Path != path || r.URL.Query().Get("query") != query {
					t.Fatalf("wrong query for %s", id)
				}
				return jsonResponse(200, `{"status":"success","data":{"resultType":"`+kind+`","result":[]}}`), nil
			})})
			if err != nil {
				t.Fatal(err)
			}
			if history {
				_, err = client.QueryRange(context.Background(), id, time.Unix(3600, 0), time.Unix(7200, 0), time.Minute)
			} else {
				_, err = client.Query(context.Background(), id, nil)
			}
			if err != nil {
				t.Fatal(err)
			}
		}
	}
}
