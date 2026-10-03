package external

import (
	"context"
	"encoding/json"
	"reflect"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/s3"
)

func (s *Service) topicOption(ctx context.Context, request Request) (cephdomain.ActionResult, error) {
	arn, name, err := topicIdentity(request)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	option, _ := request.Parameters["option"].(string)
	value, valueOK := request.Parameters["value"].(string)
	expected, expectedOK := request.Parameters["expected_value"].(string)
	status, _ := request.Parameters["expected_status"].(string)
	if !valueOK || !expectedOK || s3.ValidateTopicOption(option, value) != nil || (status != "unset" && status != "returned") {
		return cephdomain.ActionResult{}, failure("invalid_request", "supported option and current snapshot required", false)
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
	dest, destErr := topicDestination(before)
	var raw string
	_ = json.Unmarshal(dest["EndpointArgs"], &raw)
	desired, prepareErr := s3.PrepareTopicOption(raw, option, status, expected, value)
	if err != nil || destErr != nil || prepareErr != nil || before["TopicArn"] != arn || before["Name"] != name || before["User"] == "" {
		return cephdomain.ActionResult{}, failure("pre_check_failed", "topic identity, option snapshot or native replacement could not be verified; no change submitted", false)
	}
	if err := api.SetTopicOption(ctx, arn, option, value); err != nil {
		return cephdomain.ActionResult{}, failure("sns_failed", "topic option write outcome is uncertain; refresh before another change", false)
	}
	after, err := api.GetTopicAttributes(ctx, arn)
	afterDest, afterDestErr := topicDestination(after)
	var actual string
	_ = json.Unmarshal(afterDest["EndpointArgs"], &actual)
	delete(dest, "EndpointArgs")
	delete(afterDest, "EndpointArgs")
	delete(before, "EndPoint")
	delete(after, "EndPoint")
	if err != nil || afterDestErr != nil || actual != desired || !reflect.DeepEqual(before, after) || !reflect.DeepEqual(dest, afterDest) {
		return cephdomain.ActionResult{}, failure("post_check_failed", "topic option was submitted but complete attributes could not be verified; delivery settings may already have changed", false)
	}
	return cephdomain.ActionResult{}, nil
}
