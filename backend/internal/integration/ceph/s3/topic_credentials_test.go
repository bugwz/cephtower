package s3

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTopicCredentialPresenceNeverReturnsValues(t *testing.T) {
	for _, tc := range []struct{ raw, user, password string }{
		{"", "unset", "unset"},
		{"user-name=never-store&password=never-store", "set", "set"},
		{"user-name=&password", "empty", "empty"},
		{"?user%2Dname%3Dnever-store&password=%20", "set", "set"},
		{"password=one&password=two&password=", "unset", "duplicate"},
		{"user-name=&user%2Dname=never-store", "duplicate", "unset"},
		{"unknown=password%3Dnever-store", "unset", "unset"},
		{"password=never-store%26user-name%3Dnever-store", "unset", "set"},
		{"user-name=never-store&unknown=%zz", "unavailable", "unavailable"},
	} {
		got := TopicCredentialPresence(tc.raw)
		if len(got) != 2 || got[0].Field != "username" || got[1].Field != "password" || got[0].State != tc.user || got[1].State != tc.password {
			t.Fatalf("wrong presence projection: %v", got)
		}
		body, _ := json.Marshal(got)
		if strings.Contains(string(body), "never-store") {
			t.Fatal("secret persisted")
		}
	}
}
