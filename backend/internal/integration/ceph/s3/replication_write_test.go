package s3

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

type replicationTransport func(*http.Request) (*http.Response, error)

func (f replicationTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestDashboardReplicationCredentialScope(t *testing.T) {
	for _, tc := range []struct {
		tenant, owner, token string
		valid                bool
	}{
		{"", "user", "", true}, {"team", "team$user", "", true},
		{"", "team$user", "", false}, {"team", "user", "", false},
		{"team", "team$ns$user", "", true}, {"", "$ns$user", "", true},
		{"", "", "", false}, {"team", "team$", "", false},
		{"team", "team$a$b$c", "", false}, {"", "user", "token", false},
		{"", "user</ID><ID>user", "", false}, {"", "<nested/>", "", false},
		{"", "user</ID></Owner><Owner><ID>user", "", false},
		{"", " user ", "", false}, {"", "<broken", "", false},
	} {
		t.Run(tc.tenant+"/"+tc.owner+"/"+tc.token, func(t *testing.T) {
			calls := 0
			client, err := New("https://s3.test", Credentials{AccessKey: "access", SecretKey: "secret", SessionToken: tc.token}, &http.Client{Transport: replicationTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Header.Get("Authorization") == "" {
					t.Fatal("unsigned request")
				}
				body := `<ListAllMyBucketsResult><Owner><ID>` + tc.owner + `</ID></Owner><Buckets/></ListAllMyBucketsResult>`
				if calls == 1 {
					if r.Method != "GET" || r.URL.Path != "/" || r.URL.RawQuery != "" {
						t.Fatal("wrong owner request")
					}
				} else {
					if !tc.valid || calls != 2 || r.Method != "PUT" || r.URL.Path != "/"+tc.tenant+":bucket" || r.URL.RawQuery != "replication=" {
						t.Fatal("wrong replication write")
					}
					data, _ := io.ReadAll(r.Body)
					parsed, err := BucketReplication(data)
					if err != nil || len(parsed.Rules) != 1 || parsed.Rules[0].ID != "dashboard_admin_pipe" || parsed.Rules[0].DestinationBucket != "arn:aws:s3:::bucket" || parsed.Rules[0].Status != "Enabled" || *parsed.Rules[0].Priority != "0" {
						t.Fatalf("wrong document %s", data)
					}
					body = ""
				}
				return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(body))}, nil
			})})
			if err != nil {
				t.Fatal(err)
			}
			err = client.PutDashboardBucketReplication(context.Background(), tc.tenant+":bucket")
			want := 1
			if tc.valid {
				want = 2
			}
			if tc.token != "" {
				want = 0
			}
			if (err == nil) != tc.valid || calls != want {
				t.Fatalf("calls=%d err=%v", calls, err)
			}
		})
	}
}

func TestDashboardReplicationNormalizedReadback(t *testing.T) {
	for _, tenant := range []string{"", "team"} {
		body := `<ReplicationConfiguration><Role/><Rule><ID>dashboard_admin_pipe</ID><Status>Enabled</Status><Priority>0</Priority><Destination><Bucket>arn:aws:s3::` + tenant + `:bucket</Bucket></Destination></Rule></ReplicationConfiguration>`
		if !DashboardBucketReplicationMatches([]byte(body), tenant+":bucket") {
			t.Fatal("native normalization rejected")
		}
		for _, invalid := range []string{
			strings.Replace(body, "Enabled", "Disabled", 1),
			strings.Replace(body, "<Priority>0", "<Priority>1", 1),
			strings.Replace(body, "dashboard_admin_pipe", "other", 1),
			strings.Replace(body, ":bucket<", ":different<", 1),
			strings.Replace(body, "<Role/>", "<Role/><Role/>", 1),
			strings.Replace(body, "</Rule>", "<Filter><Prefix>objects/</Prefix></Filter></Rule>", 1),
			strings.Replace(body, "</Destination>", "<Zone>other</Zone></Destination>", 1),
			strings.Replace(body, "<Role/>", "<Role>retained</Role>", 1),
			strings.Replace(body, "</Rule>", "<Status>Enabled</Status></Rule>", 1),
			body + body, "broken",
		} {
			if DashboardBucketReplicationMatches([]byte(invalid), tenant+":bucket") {
				t.Fatalf("false positive %s", invalid)
			}
		}
	}
}
