package s3

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestBucketMFASecretHeader(t *testing.T) {
	for _, scheme := range []string{"https", "http"} {
		calls := 0
		client, err := New(scheme+"://s3.example.test", Credentials{AccessKey: "access", SecretKey: "secret"}, &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			body, _ := io.ReadAll(r.Body)
			if r.Method != "PUT" || r.URL.RawQuery != "versioning=" || r.Header.Get("X-Amz-Mfa") != "device:serial 001234" || !strings.Contains(r.Header.Get("Authorization"), "x-amz-mfa") || strings.Contains(string(body), "001234") || strings.Contains(r.URL.String(), "001234") || !strings.Contains(string(body), "<MfaDelete>Enabled</MfaDelete>") {
				t.Fatal("unsafe or unsigned MFA transport")
			}
			return &http.Response{StatusCode: 403, Header: http.Header{}, Body: io.NopCloser(strings.NewReader("device:serial 001234"))}, nil
		})})
		if err != nil {
			t.Fatal(err)
		}
		err = client.PutBucketMFA(context.Background(), ":bucket", "Enabled", "Enabled", "device:serial", "001234")
		if err == nil || strings.Contains(err.Error(), "001234") || strings.Contains(err.Error(), "device:serial") {
			t.Fatal("raw MFA secret leaked")
		}
		if scheme == "https" && calls != 1 || scheme == "http" && calls != 0 {
			t.Fatal("HTTPS guard missing")
		}
	}
	for _, args := range [][4]string{{"bad", "Enabled", "device", "123456"}, {"Enabled", "bad", "device", "123456"}, {"Enabled", "Enabled", "device\r\nx", "123456"}, {"Enabled", "Enabled", "device x", "123456"}, {"Enabled", "Enabled", "device", ""}, {"Enabled", "Enabled", "device", "123 456"}} {
		if ValidateBucketMFA(args[0], args[1], args[2], args[3]) == nil {
			t.Fatal("invalid MFA request accepted")
		}
	}
}

func TestVersioningMFAProjection(t *testing.T) {
	for _, mfa := range []string{"", "Enabled", "Disabled"} {
		body := "<VersioningConfiguration><Status>Enabled</Status>"
		if mfa != "" {
			body += "<MfaDelete>" + mfa + "</MfaDelete>"
		}
		got, err := BucketVersioning([]byte(body + "</VersioningConfiguration>"))
		if err != nil || got.Status != "Enabled" || (got.MFADelete == nil) != (mfa == "") || got.MFADelete != nil && *got.MFADelete != mfa {
			t.Fatalf("lost MFA state: %+v %v", got, err)
		}
	}
	empty, err := BucketVersioning([]byte("<VersioningConfiguration/>"))
	if err != nil || empty.Status != "" || empty.MFADelete != nil {
		t.Fatal("invented never-versioned state")
	}
}

func TestMFACredentialsNeverFollowRedirects(t *testing.T) {
	for _, target := range []string{"https://other.example.test/bucket", "http://s3.example.test/bucket", "https://s3.example.test/other"} {
		calls := 0
		httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			if calls > 1 {
				t.Fatal("MFA credentials replayed")
			}
			return &http.Response{StatusCode: 307, Header: http.Header{"Location": {target}}, Body: io.NopCloser(strings.NewReader(""))}, nil
		})}
		client, err := New("https://s3.example.test", Credentials{AccessKey: "access", SecretKey: "secret"}, httpClient)
		if err != nil {
			t.Fatal(err)
		}
		if client.PutBucketMFA(context.Background(), ":bucket", "Enabled", "Enabled", "serial", "001234") == nil || calls != 1 {
			t.Fatal("redirect counted as successful write")
		}
		if httpClient.CheckRedirect != nil {
			t.Fatal("shared client modified")
		}
	}
}
