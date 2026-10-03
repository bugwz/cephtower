package external

import (
	"context"
	"reflect"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/s3"
)

func deleteBucketNotification(ctx context.Context, api *s3.Client, bucket string, parameters map[string]any) (cephdomain.ActionResult, error) {
	mode, modeOK := parameters["mode"].(string)
	id, idOK := parameters["notification_id"].(string)
	expected, expectedOK := parameters["expected_document"].(string)
	if !modeOK || !idOK || !expectedOK || expected == "" || s3.ValidateNotificationDeletion(mode, id) != nil {
		return cephdomain.ActionResult{}, failure("invalid_request", "explicit deletion mode, notification ID and full snapshot required", false)
	}
	before, _, err := api.GetBucketConfiguration(ctx, bucket, "notification")
	if err != nil || string(before) != expected {
		return cephdomain.ActionResult{}, failure("pre_check_failed", "notification snapshot unavailable or changed; no deletion submitted", false)
	}
	rules, err := s3.BucketNotifications(before)
	if err != nil || len(rules) == 0 {
		return cephdomain.ActionResult{}, failure("pre_check_failed", "no confirmed native notifications to delete", false)
	}
	want := []s3.BucketNotification{}
	if mode == "single" {
		matches := 0
		for _, rule := range rules {
			if rule.ID == id {
				matches++
			} else {
				want = append(want, rule)
			}
		}
		if matches != 1 {
			return cephdomain.ActionResult{}, failure("pre_check_failed", "notification ID is missing or ambiguous; no deletion submitted", false)
		}
	}
	if err := api.DeleteBucketNotification(ctx, bucket, mode, id); err != nil {
		return cephdomain.ActionResult{}, failure("s3_failed", "notification deletion failed and may have partially applied; refresh before another change; no automatic retry", false)
	}
	after, _, readErr := api.GetBucketConfiguration(ctx, bucket, "notification")
	actual, parseErr := s3.BucketNotifications(after)
	if readErr != nil || parseErr != nil || !reflect.DeepEqual(want, actual) {
		return cephdomain.ActionResult{}, failure("post_check_failed", "notification deletion was submitted but remaining rules could not be verified; no automatic rollback or retry", false)
	}
	return cephdomain.ActionResult{}, nil
}
