package external

import (
	"context"
	"encoding/json"
	"reflect"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/s3"
)

func (s *Service) topicEndpoint(ctx context.Context, request Request) (cephdomain.ActionResult, error) {
	arn, name, err := topicIdentity(request)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	value, valueOK := request.Parameters["endpoint_secret"].(string)
	expected, expectedOK := request.Parameters["expected_endpoint"].(string)
	redacted, redactedOK := request.Parameters["expected_redacted"].(bool)
	stored, storedOK := request.Parameters["expected_stored_secret"].(bool)
	hasSecret, err := s3.ValidateTopicEndpoint(value)
	visible, hidden, valid := s3.RedactTopicEndpoint(expected)
	if err != nil || !valueOK || !expectedOK || !redactedOK || !storedOK || !valid || hidden || visible != expected {
		return cephdomain.ActionResult{}, failure("invalid_request", "valid endpoint and redacted current snapshot required", false)
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
	var old string
	var oldStored bool
	_ = json.Unmarshal(dest["EndpointAddress"], &old)
	_ = json.Unmarshal(dest["HasStoredSecret"], &oldStored)
	current, currentRedacted, currentValid := s3.RedactTopicEndpoint(old)
	if err != nil || destErr != nil || before["TopicArn"] != arn || before["Name"] != name || before["User"] == "" || !currentValid || current != expected || currentRedacted != redacted || oldStored != stored || old == value {
		return cephdomain.ActionResult{}, failure("pre_check_failed", "current SNS topic or visible endpoint snapshot could not be verified, or endpoint is unchanged; no change submitted", false)
	}
	if err := api.SetTopicEndpoint(ctx, arn, value); err != nil {
		return cephdomain.ActionResult{}, failure("sns_failed", "topic endpoint write outcome is uncertain; refresh before another change", false)
	}
	after, err := api.GetTopicAttributes(ctx, arn)
	afterDest, afterDestErr := topicDestination(after)
	var actual string
	_ = json.Unmarshal(afterDest["EndpointAddress"], &actual)
	// RGW only sets stored_secret when credentials are present; it does not clear
	// the previous marker when replacing or removing the endpoint URL.
	dest["HasStoredSecret"], _ = json.Marshal(oldStored || hasSecret)
	delete(dest, "EndpointAddress")
	delete(afterDest, "EndpointAddress")
	delete(before, "EndPoint")
	delete(after, "EndPoint")
	if err != nil || afterDestErr != nil || actual != value || !reflect.DeepEqual(before, after) || !reflect.DeepEqual(dest, afterDest) {
		return cephdomain.ActionResult{}, failure("post_check_failed", "topic endpoint was submitted but complete attributes could not be verified; delivery or queue changes may already have occurred", false)
	}
	return cephdomain.ActionResult{}, nil
}
