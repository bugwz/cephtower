package s3

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"unicode/utf8"
)

// Empty IDs have destructive native semantics, so callers must select all
// explicitly. Never trim an ID or turn a malformed single delete into delete-all.
func ValidateNotificationDeletion(mode, id string) error {
	if mode == "all" && id == "" {
		return nil
	}
	if mode != "single" || id == "" || !utf8.ValidString(id) || strings.ContainsAny(id, "\x00\r\n") {
		return fmt.Errorf("select single with an exact nonempty notification ID, or all with an empty ID")
	}
	return nil
}

func (c *Client) DeleteBucketNotification(ctx context.Context, bucket, mode, id string) error {
	if err := ValidateNotificationDeletion(mode, id); err != nil {
		return err
	}
	_, _, err := c.request(ctx, http.MethodDelete, bucket, url.Values{"notification": {id}}, nil)
	return err
}
