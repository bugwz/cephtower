package iscsi

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

func TestISCSIClientUsesTLSBasicAuthAndTypedTarget(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "api" || password != "secret" {
			t.Fatal("basic auth missing")
		}
		if r.URL.Path == "/api/gateway" {
			return jsonResponse(http.StatusOK, `{"status":"UP"}`), nil
		}
		if r.URL.Path != "/api/target/iqn.2026-07.example:test" || r.Method != http.MethodPut {
			t.Fatalf("unexpected request %s %s", r.Method, r.URL.EscapedPath())
		}
		return jsonResponse(http.StatusOK, ""), nil
	})}
	client, err := New("https://iscsi.example.test", "api", "secret", httpClient)
	if err != nil {
		t.Fatal(err)
	}
	health, err := client.Health(context.Background())
	if err != nil || health.Status != "UP" {
		t.Fatalf("health=%#v err=%v", health, err)
	}
	if err := client.ApplyTarget(context.Background(), Target{IQN: "iqn.2026-07.example:test"}); err != nil {
		t.Fatal(err)
	}
}

func TestISCSIClientRequiresHTTPS(t *testing.T) {
	if _, err := New("http://gateway.example.test", "", "", nil); err == nil {
		t.Fatal("insecure endpoint accepted")
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }
func jsonResponse(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: make(http.Header), Body: io.NopCloser(strings.NewReader(body))}
}
