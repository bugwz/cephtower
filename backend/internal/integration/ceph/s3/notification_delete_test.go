package s3

import (
	"context"
	"testing"
)

func TestNotificationDeletionRequiresExplicitScope(t *testing.T) {
	for _, test := range []struct {
		mode, id string
		valid    bool
	}{
		{"all", "", true}, {"single", " exact &+%/中 ", true},
		{"", "", false}, {"single", "", false}, {"all", "id", false},
		{"single", "a\x00b", false}, {"single", "a\nb", false}, {"single", "a\rb", false}, {"single", string([]byte{255}), false},
	} {
		if (ValidateNotificationDeletion(test.mode, test.id) == nil) != test.valid {
			t.Fatalf("invalid scope classification: %+v", test)
		}
		if !test.valid {
			// A nil client proves invalid scopes are rejected before any request.
			var client *Client
			if client.DeleteBucketNotification(context.Background(), ":bucket", test.mode, test.id) == nil {
				t.Fatal("invalid scope reached request")
			}
		}
	}
}
