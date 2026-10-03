package external

import (
	"context"
	"encoding/json"
	"reflect"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/s3"
)

func (s *Service) topicAttribute(ctx context.Context, request Request) (cephdomain.ActionResult, error) {
	arn, name, err := topicIdentity(request)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	attribute, _ := request.Parameters["attribute"].(string)
	value, valueOK := request.Parameters["value"].(string)
	expected, expectedOK := request.Parameters["expected_value"].(string)
	if _, err := s3.TopicAttributeWireValue(attribute, value); err != nil || !valueOK || !expectedOK || value == expected {
		return cephdomain.ActionResult{}, failure("invalid_request", "changed supported attribute and current snapshot required", false)
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
	if err != nil || destErr != nil || before["TopicArn"] != arn || before["Name"] != name || before["User"] == "" || !topicAttributeMatches(before, dest, attribute, expected) {
		return cephdomain.ActionResult{}, failure("pre_check_failed", "current SNS topic identity or attribute could not be verified; no change submitted", false)
	}
	if err := api.SetTopicAttribute(ctx, arn, attribute, value); err != nil {
		return cephdomain.ActionResult{}, failure("sns_failed", "topic attribute write outcome is uncertain; refresh before another change", false)
	}
	after, err := api.GetTopicAttributes(ctx, arn)
	afterDest, afterDestErr := topicDestination(after)
	if attribute == "OpaqueData" {
		before["OpaqueData"] = value
	} else {
		var raw []byte
		if attribute == "persistent" {
			raw = []byte(value)
		} else {
			raw, _ = json.Marshal(value)
		}
		dest[topicDestinationField(attribute)] = raw
	}
	// The native JSON formatter is stable. Compare complete endpoint fields while
	// allowing only the selected field to differ; never return endpoint secrets.
	delete(before, "EndPoint")
	delete(after, "EndPoint")
	if err != nil || afterDestErr != nil || !reflect.DeepEqual(before, after) || !reflect.DeepEqual(dest, afterDest) {
		return cephdomain.ActionResult{}, failure("post_check_failed", "topic attribute was submitted but complete attributes could not be verified; queue changes may already have occurred", false)
	}
	return cephdomain.ActionResult{}, nil
}

func topicDestination(attributes map[string]string) (map[string]json.RawMessage, error) {
	var dest map[string]json.RawMessage
	err := json.Unmarshal([]byte(attributes["EndPoint"]), &dest)
	if err != nil || dest == nil {
		return nil, failure("invalid_response", "SNS destination unavailable", false)
	}
	// Require the current native fields, including secrets that remain in memory.
	for _, key := range []string{"EndpointAddress", "EndpointArgs", "EndpointTopic", "TimeToLive", "MaxRetries", "RetrySleepDuration"} {
		var value string
		if raw, ok := dest[key]; !ok || string(raw) == "null" || json.Unmarshal(raw, &value) != nil {
			return nil, failure("invalid_response", "SNS destination incomplete", false)
		}
	}
	for _, key := range []string{"Persistent", "HasStoredSecret"} {
		var value bool
		if raw, ok := dest[key]; !ok || string(raw) == "null" || json.Unmarshal(raw, &value) != nil {
			return nil, failure("invalid_response", "SNS destination incomplete", false)
		}
	}
	return dest, nil
}

func topicDestinationField(attribute string) string {
	return map[string]string{"persistent": "Persistent", "time_to_live": "TimeToLive", "max_retries": "MaxRetries", "retry_sleep_duration": "RetrySleepDuration"}[attribute]
}
func topicAttributeMatches(attributes map[string]string, dest map[string]json.RawMessage, attribute, expected string) bool {
	if attribute == "OpaqueData" {
		return attributes["OpaqueData"] == expected
	}
	raw := dest[topicDestinationField(attribute)]
	if attribute == "persistent" {
		return (expected == "true" || expected == "false") && string(raw) == expected
	}
	var value string
	return json.Unmarshal(raw, &value) == nil && value == expected
}
