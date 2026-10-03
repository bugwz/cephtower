package external

import (
	"context"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/s3"
)

func setBucketMFA(ctx context.Context, api *s3.Client, bucket string, parameters map[string]any) (cephdomain.ActionResult, error) {
	status, _ := parameters["status"].(string)
	mfa, _ := parameters["mfa_delete"].(string)
	serial, _ := parameters["mfa_serial_secret"].(string)
	token, _ := parameters["mfa_token"].(string)
	expected, _ := parameters["expected_document"].(string)
	if expected == "" || s3.ValidateBucketMFA(status, mfa, serial, token) != nil {
		return cephdomain.ActionResult{}, failure("invalid_request", "full snapshot, explicit versioning/MFA states and valid MFA credentials required", false)
	}
	before, _, err := api.GetBucketConfiguration(ctx, bucket, "versioning")
	current, parseErr := s3.BucketVersioning(before)
	if err != nil || parseErr != nil || string(before) != expected || current.Status != "" && current.MFADelete == nil {
		return cephdomain.ActionResult{}, failure("pre_check_failed", "versioning snapshot unavailable, incomplete or changed; no write submitted", false)
	}
	if current.Status == status && current.MFADelete != nil && *current.MFADelete == mfa {
		return cephdomain.ActionResult{}, failure("invalid_request", "versioning and MFA state are unchanged", false)
	}
	if err := api.PutBucketMFA(ctx, bucket, status, mfa, serial, token); err != nil {
		return cephdomain.ActionResult{}, failure("s3_failed", "MFA versioning write failed and may have applied; verify configuration and obtain a fresh code before another attempt; no automatic retry", false)
	}
	after, _, readErr := api.GetBucketConfiguration(ctx, bucket, "versioning")
	actual, parseErr := s3.BucketVersioning(after)
	if readErr != nil || parseErr != nil || actual.Status != status || actual.MFADelete == nil || *actual.MFADelete != mfa {
		return cephdomain.ActionResult{}, failure("post_check_failed", "MFA versioning write submitted but both states could not be verified; no automatic retry or rollback", false)
	}
	return cephdomain.ActionResult{}, nil
}
