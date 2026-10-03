package s3

import "testing"

func TestPrepareTopicOption(t *testing.T) {
	for _, tc := range []struct {
		raw, name, status, old, value, want string
		fail                                bool
	}{
		{"", "verify-ssl", "unset", "", "false", "&verify-ssl=false", false},
		{"password=secret&verify-ssl=true&unknown=kept", "verify-ssl", "returned", "true", "false", "password=secret&verify-ssl=false&unknown=kept", false},
		{"ca-location=%2Fetc%2Fold.pem", "ca-location", "returned", "/etc/old.pem", "/etc/new.pem", "ca-location=/etc/new.pem", false},
		{"amqp-exchange=old", "amqp-exchange", "returned", "old", "", "amqp-exchange=", false},
		{"verify-ssl=true&verify-ssl=false", "verify-ssl", "returned", "true", "false", "", true},
		{"password=verify-ssl-secret&verify-ssl=true", "verify-ssl", "returned", "true", "false", "", true},
		{"x-verify-ssl=true", "verify-ssl", "unset", "", "false", "", true},
		{"verify-ssl%3Dtrue", "verify-ssl", "returned", "true", "false", "", true},
		{"verify%2Dssl=true", "verify-ssl", "returned", "true", "false", "", true},
		{"verify-ssl=true", "verify-ssl", "returned", "false", "true", "", true},
		{"verify-ssl=true", "verify-ssl", "returned", "true", "true", "", true},
		{"password=%zz", "verify-ssl", "unset", "", "true", "", true},
		{"", "password", "unset", "", "newsecret", "", true},
		{"", "kafka-brokers", "unset", "", "broker:9092", "", true},
		{"", "ca-location", "unset", "", "/path&password=injection", "", true},
		{"", "amqp-exchange", "unset", "", "a%26password=injection", "", true},
	} {
		t.Run(tc.raw+tc.name, func(t *testing.T) {
			got, err := PrepareTopicOption(tc.raw, tc.name, tc.status, tc.old, tc.value)
			if tc.fail {
				if err == nil {
					t.Fatal("unsafe native update accepted")
				}
			} else if err != nil || got != tc.want {
				t.Fatalf("unexpected native result %q %v", got, err)
			}
		})
	}
}
