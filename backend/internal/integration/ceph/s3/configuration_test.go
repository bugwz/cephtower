package s3

import (
	"context"
	"crypto/md5"
	"encoding/base64"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestBucketConfigurationWireFormats(t *testing.T) {
	for kind, body := range map[string]string{"policy": `{"Statement":[],"large":9007199254740993}`, "cors": `<CORSConfiguration><CORSRule><AllowedOrigin>*</AllowedOrigin><AllowedMethod>GET</AllowedMethod></CORSRule></CORSConfiguration>`, "lifecycle": `<LifecycleConfiguration xmlns="http://s3.amazonaws.com/doc/2006-03-01/"><Rule><ID>expire</ID><Status>Enabled</Status><Filter/><Expiration><Days>30</Days></Expiration></Rule></LifecycleConfiguration>`, "encryption": `<ServerSideEncryptionConfiguration><Rule><ApplyServerSideEncryptionByDefault><SSEAlgorithm>AES256</SSEAlgorithm></ApplyServerSideEncryptionByDefault></Rule></ServerSideEncryptionConfiguration>`, "versioning": `<VersioningConfiguration><Status>Enabled</Status></VersioningConfiguration>`} {
		calls := 0
		client, _ := New("https://s3.example.test", Credentials{AccessKey: "access", SecretKey: "secret"}, &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			data, _ := io.ReadAll(r.Body)
			digest := md5.Sum([]byte(body))
			contentType := "application/xml"
			if kind == "policy" {
				contentType = "application/json"
			}
			if string(data) != body || r.Method != "PUT" || r.URL.RawQuery != kind+"=" || r.Header.Get("Content-Type") != contentType || r.Header.Get("Content-MD5") != base64.StdEncoding.EncodeToString(digest[:]) || r.Header.Get("Authorization") == "" {
				t.Fatalf("incorrect %s request: %+v %s", kind, r, data)
			}
			return &http.Response{StatusCode: 200, Body: io.NopCloser(strings.NewReader("")), Header: http.Header{}}, nil
		})})
		if err := client.PutBucketConfiguration(context.Background(), "bucket", kind, []byte(body)); err != nil || calls != 1 {
			t.Fatalf("%s: %v calls=%d", kind, err, calls)
		}
	}
}

func TestInvalidBucketConfigurationDoesNotSend(t *testing.T) {
	client, _ := New("https://s3.example.test", Credentials{AccessKey: "access", SecretKey: "secret"}, &http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) { t.Fatal("invalid configuration sent"); return nil, nil })})
	for _, tc := range []struct{ kind, body string }{
		{"policy", "null"}, {"policy", "[]"}, {"policy", "{}{}"}, {"policy", ""}, {"cors", "{}"}, {"cors", "<Wrong/>"}, {"cors", "<CORSConfiguration>"}, {"cors", "<CORSConfiguration/><CORSConfiguration/>"}, {"cors", "text<CORSConfiguration/>"}, {"cors", "<!DOCTYPE x><CORSConfiguration/>"}, {"cors", "<?other test?><CORSConfiguration/>"}, {"cors", "<CORSConfiguration>&unknown;</CORSConfiguration>"}, {"unknown", "{}"}, {"encryption", "<LifecycleConfiguration/>"},
	} {
		if err := client.PutBucketConfiguration(context.Background(), "bucket", tc.kind, []byte(tc.body)); err == nil {
			t.Fatalf("accepted %s %s", tc.kind, tc.body)
		}
	}
}
