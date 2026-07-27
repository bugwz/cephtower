package s3

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) { return f(request) }

func TestSignedBucketRequests(t *testing.T) {
	requests := make(chan *http.Request, 3)
	bodies := make(chan string, 3)
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		body, _ := io.ReadAll(r.Body)
		requests <- r.Clone(context.Background())
		bodies <- string(body)
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("")), Header: make(http.Header), Request: r}, nil
	})}
	client, err := New("https://s3.example.test", Credentials{AccessKey: "access", SecretKey: "secret", Region: "us-east-1"}, httpClient)
	if err != nil {
		t.Fatal(err)
	}
	client.now = func() time.Time { return time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC) }
	if err := client.CreateBucket(context.Background(), "bucket-one"); err != nil {
		t.Fatal(err)
	}
	if err := client.PutBucketConfiguration(context.Background(), "bucket-one", "policy", []byte(`{"Version":"2012-10-17"}`)); err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteBucket(context.Background(), "bucket-one"); err != nil {
		t.Fatal(err)
	}
	for index := 0; index < 3; index++ {
		request := <-requests
		body := <-bodies
		if request.URL.Path != "/bucket-one" || !strings.HasPrefix(request.Header.Get("Authorization"), "AWS4-HMAC-SHA256 Credential=access/") {
			t.Fatalf("request = %s %s auth=%q", request.Method, request.URL.String(), request.Header.Get("Authorization"))
		}
		if request.Method == http.MethodPut && request.URL.Query().Has("policy") && body == "" {
			t.Fatal("policy request body is empty")
		}
	}
}

func TestS3ErrorIsBounded(t *testing.T) {
	httpClient := &http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: http.StatusBadGateway, Body: io.NopCloser(strings.NewReader(strings.Repeat("x", 64<<10))), Header: make(http.Header), Request: r}, nil
	})}
	client, err := New("https://s3.example.test", Credentials{AccessKey: "access", SecretKey: "secret"}, httpClient)
	if err != nil {
		t.Fatal(err)
	}
	err = client.DeleteBucket(context.Background(), "bucket")
	if err == nil || len(err.Error()) > 34<<10 {
		t.Fatalf("error length = %d, error = %v", len(err.Error()), err)
	}
}
