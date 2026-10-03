package s3

import (
	"strings"
	"testing"
)

func TestBucketNotificationDocument(t *testing.T) {
	rule := BucketNotification{ID: " <&中 ", Topic: "arn:aws:sns:east:team:topic", Events: []string{}, Filters: []BucketNotificationFilter{{"S3Tags", "a", ""}, {"S3Key", "regex", "a.*"}, {"S3Key", "prefix", ""}, {"S3Metadata", "x-amz-meta-key", "<&\r"}}}
	body, want, err := BucketNotificationDocument(rule)
	if err != nil {
		t.Fatal(err)
	}
	parsed, err := BucketNotifications(body)
	if err != nil || len(parsed) != 1 || parsed[0].ID != rule.ID || len(want.Filters) != 3 || len(want.Events) != 2 || want.Events[0] != "s3:ObjectCreated:*" || !strings.Contains(string(body), "&#xD;") {
		t.Fatalf("bad XML projection: %s %+v %v", body, want, err)
	}
	for _, event := range []string{"s3:ObjectCreated:Post", "s3:ObjectLifecycle:Expiration:Noncurrent", "s3:ObjectLifecycle:Expiration:AbortMultipartUpload", "s3:Replication:*"} {
		rule.Events = []string{event}
		_, projected, err := BucketNotificationDocument(rule)
		if err != nil || len(projected.Events) != 1 {
			t.Fatal(err)
		}
		if event == "s3:ObjectLifecycle:Expiration:AbortMultipartUpload" && projected.Events[0] != "s3:ObjectLifecycle:Expiration:AbortMPU" {
			t.Fatal("native event normalization lost")
		}
	}
	for _, event := range []string{"s3:ObjectRestore:*", "s3:UnknownEvent", "s3:ObjectLifecycle:Expiration:AbortMPU", ""} {
		rule.Events = []string{event}
		if _, _, err := BucketNotificationDocument(rule); err == nil {
			t.Fatal("unsupported native event accepted")
		}
	}
	rule.Events = []string{}
	for _, patch := range []func(*BucketNotification){
		func(r *BucketNotification) { r.ID = "" }, func(r *BucketNotification) { r.ID = "\x00" }, func(r *BucketNotification) { r.Topic = "arn:aws:s3:::bucket" }, func(r *BucketNotification) { r.Events = nil }, func(r *BucketNotification) { r.Filters = nil },
		func(r *BucketNotification) { r.Filters = []BucketNotificationFilter{{"S3Key", "unknown", "x"}} },
		func(r *BucketNotification) {
			r.Filters = []BucketNotificationFilter{{"S3Tags", "key", "a"}, {"S3Tags", "key", "b"}}
		},
		func(r *BucketNotification) { r.Filters = []BucketNotificationFilter{{"S3Tags", "key", "\x01"}} },
	} {
		copy := rule
		patch(&copy)
		if _, _, err := BucketNotificationDocument(copy); err == nil {
			t.Fatal("invalid rule accepted")
		}
	}
	a := []BucketNotification{want, {ID: "other", Topic: "arn:aws:sns:east::other", Events: []string{"future"}, Filters: []BucketNotificationFilter{}}}
	b := []BucketNotification{a[1], a[0]}
	b[1].Filters = []BucketNotificationFilter{want.Filters[2], want.Filters[1], want.Filters[0]}
	if !SameBucketNotifications(a, b) {
		t.Fatal("native ordering treated as data change")
	}
	b[1].Filters = append(b[1].Filters, want.Filters[0])
	if SameBucketNotifications(a, b) || SameBucketNotifications(a, append(a, a[0])) {
		t.Fatal("duplicate data lost")
	}
}
