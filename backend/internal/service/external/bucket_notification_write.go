package external

import (
	"bytes"
	"context"
	"encoding/json"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/s3"
)

func setBucketNotification(ctx context.Context, api *s3.Client, bucket string, parameters map[string]any) (cephdomain.ActionResult, error) {
	mode, _ := parameters["mode"].(string)
	expected, _ := parameters["expected_document"].(string)
	encoded, err := json.Marshal(parameters["rule"])
	var rule s3.BucketNotification
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err != nil || decoder.Decode(&rule) != nil || expected == "" || (mode != "create" && mode != "edit") {
		return cephdomain.ActionResult{}, failure("invalid_request", "explicit mode, full snapshot and notification rule required", false)
	}
	_, desired, err := s3.BucketNotificationDocument(rule)
	if err != nil {
		return cephdomain.ActionResult{}, failure("invalid_request", err.Error(), false)
	}
	before, _, err := api.GetBucketConfiguration(ctx, bucket, "notification")
	if err != nil || string(before) != expected {
		return cephdomain.ActionResult{}, failure("pre_check_failed", "notification snapshot unavailable or changed; no write submitted", false)
	}
	rules, err := s3.BucketNotifications(before)
	if err != nil {
		return cephdomain.ActionResult{}, failure("pre_check_failed", "notification snapshot malformed; no write submitted", false)
	}
	remaining := []s3.BucketNotification{}
	matches, oldTopic := 0, ""
	topicName, _ := s3.NotificationTopicName(rule.Topic)
	for _, existing := range rules {
		if existing.ID == rule.ID {
			matches++
			oldTopic = existing.Topic
			continue
		}
		name, parseErr := s3.NotificationTopicName(existing.Topic)
		if parseErr != nil || existing.ID+"_"+name == rule.ID+"_"+topicName {
			return cephdomain.ActionResult{}, failure("pre_check_failed", "cannot exclude native notification storage-key collision; no write submitted", false)
		}
		remaining = append(remaining, existing)
	}
	if mode == "create" && matches != 0 || mode == "edit" && matches != 1 {
		return cephdomain.ActionResult{}, failure("pre_check_failed", "notification ID already exists, is absent or ambiguous; no write submitted", false)
	}
	// Verify exact destination before a topic replacement can remove the old rule.
	// This requires GetTopicAttributes permission in addition to native Publish.
	attributes, err := api.GetTopicAttributes(ctx, rule.Topic)
	if err != nil || attributes["TopicArn"] != rule.Topic || attributes["Name"] != topicName {
		return cephdomain.ActionResult{}, failure("pre_check_failed", "destination topic identity unavailable; no write submitted", false)
	}
	oldName := ""
	if mode == "edit" {
		oldName, err = s3.NotificationTopicName(oldTopic)
		if err != nil {
			return cephdomain.ActionResult{}, failure("pre_check_failed", "existing topic identity invalid; no write submitted", false)
		}
	}
	if mode == "edit" && oldName != topicName {
		if err := api.DeleteBucketNotification(ctx, bucket, "single", rule.ID); err != nil {
			return cephdomain.ActionResult{}, failure("s3_failed", "old notification removal failed and may have applied; no automatic retry or rollback", false)
		}
		current, _, readErr := api.GetBucketConfiguration(ctx, bucket, "notification")
		actual, parseErr := s3.BucketNotifications(current)
		if readErr != nil || parseErr != nil || !s3.SameBucketNotifications(remaining, actual) {
			return cephdomain.ActionResult{}, failure("post_check_failed", "old notification removal could not be verified; replacement not submitted; no automatic rollback", false)
		}
	}
	if err := api.PutBucketNotification(ctx, bucket, rule); err != nil {
		return cephdomain.ActionResult{}, failure("s3_failed", "notification write failed and may have applied; a replacement may have already removed the old rule; no automatic retry or rollback", false)
	}
	after, _, readErr := api.GetBucketConfiguration(ctx, bucket, "notification")
	actual, parseErr := s3.BucketNotifications(after)
	if readErr != nil || parseErr != nil || !s3.SameBucketNotifications(append(remaining, desired), actual) {
		return cephdomain.ActionResult{}, failure("post_check_failed", "notification write submitted but full remaining configuration could not be verified; no automatic retry or rollback", false)
	}
	return cephdomain.ActionResult{}, nil
}
