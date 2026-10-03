package s3

import (
	"net/url"
	"testing"
)

func TestS3CanonicalQueryUsesSigV4EscapingAndOrdering(t *testing.T) {
	query := url.Values{"notification": {" a&+%/中 "}, "z": {"z", " "}, "中": {""}, "a": {""}, "a ": {""}}
	const want = "%E4%B8%AD=&a=&a%20=&notification=%20a%26%2B%25%2F%E4%B8%AD%20&z=%20&z=z"
	if got := canonicalQuery(query); got != want {
		t.Fatalf("canonical query %q want %q", got, want)
	}
	if got, err := url.ParseQuery(canonicalQuery(query)); err != nil || got.Get("notification") != query.Get("notification") {
		t.Fatal("notification ID changed")
	}
}
