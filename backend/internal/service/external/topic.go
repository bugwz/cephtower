package external

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/s3"
)

func topicIdentity(request Request) (string, string, error) {
	p := request.Parameters
	id, _ := p["topic_id"].(string)
	arn, _ := p["topic_arn"].(string)
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(id)
	key := string(decoded)
	scope, name, scoped := strings.Cut(key, ":")
	parts := strings.SplitN(arn, ":", 6)
	if err != nil || !utf8.Valid(decoded) || base64.RawURLEncoding.EncodeToString(decoded) != id || !scoped || name == "" || strings.IndexFunc(key, unicode.IsControl) >= 0 || len(parts) != 6 || parts[0] != "arn" || parts[1] != "aws" || parts[2] != "sns" || parts[3] == "" || parts[4] != scope || parts[5] != name || request.ResourceKey != "rgw/topic/"+id {
		return "", "", failure("invalid_request", "topic identity and ARN must agree", false)
	}
	return arn, name, nil
}
func (s *Service) topicPolicy(ctx context.Context, request Request) (cephdomain.ActionResult, error) {
	arn, name, err := topicIdentity(request)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	p := request.Parameters
	expected, expectedOK := p["expected_policy"].(string)
	policy, policyOK := p["policy"].(string)
	if !expectedOK || !policyOK || expected == policy {
		return cephdomain.ActionResult{}, failure("invalid_request", "changed policy and current snapshot required", false)
	}
	if policy != "" {
		if err := s3.ValidateBucketConfiguration("policy", []byte(policy)); err != nil {
			return cephdomain.ActionResult{}, failure("invalid_request", "topic policy must be a JSON object or empty to clear", false)
		}
	}
	endpoint, credential, client, err := s.httpClient(ctx, request.ClusterID, "s3")
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	api, err := s3.New(endpoint.URL, s3.Credentials{AccessKey: credential.AccessKey, SecretKey: credential.SecretKey, SessionToken: credential.SessionToken, Region: credential.Region}, client)
	if err != nil {
		return cephdomain.ActionResult{}, failure("invalid_credential", "SNS credentials or endpoint invalid", false)
	}
	before, err := api.GetTopicAttributes(ctx, arn)
	var dest map[string]json.RawMessage
	if err != nil || before["TopicArn"] != arn || before["Name"] != name || before["User"] == "" || before["Policy"] != expected || json.Unmarshal([]byte(before["EndPoint"]), &dest) != nil || dest == nil {
		return cephdomain.ActionResult{}, failure("pre_check_failed", "current SNS topic scope or policy could not be verified; no change submitted", false)
	}
	if err := api.SetTopicPolicy(ctx, arn, policy); err != nil {
		return cephdomain.ActionResult{}, failure("sns_failed", "topic policy write outcome is uncertain; refresh before another change", false)
	}
	after, err := api.GetTopicAttributes(ctx, arn)
	before["Policy"] = policy
	if err != nil || !reflect.DeepEqual(before, after) {
		return cephdomain.ActionResult{}, failure("post_check_failed", "topic policy was submitted but complete attributes could not be verified; access may already have changed", false)
	}
	return cephdomain.ActionResult{}, nil
}
