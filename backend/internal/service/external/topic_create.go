package external

import (
	"context"
	"encoding/json"
	"strings"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/s3"
)

func (s *Service) topicCreate(ctx context.Context, request Request) (cephdomain.ActionResult, error) {
	arn, name, err := topicIdentity(request)
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	p := request.Parameters
	text := func(key string) string { value, _ := p[key].(string); return value }
	persistent, persistentOK := p["persistent"].(bool)
	input := s3.TopicCreate{Name: name, Endpoint: text("endpoint_secret"), OpaqueData: text("opaque_data"), Policy: text("policy"), Persistent: persistent, TTL: text("time_to_live"), MaxRetries: text("max_retries"), RetrySleep: text("retry_sleep_duration"), Options: map[string]string{}}
	options, optionsOK := p["options"].(map[string]any)
	for key, value := range options {
		v, ok := value.(string)
		if !ok {
			return cephdomain.ActionResult{}, failure("invalid_request", "topic options must be strings", false)
		}
		input.Options[key] = v
	}
	if !persistentOK || !optionsOK || input.Validate() != nil || text("owner_uid") == "" {
		return cephdomain.ActionResult{}, failure("invalid_request", "complete topic creation configuration required", false)
	}
	endpoint, credential, client, err := s.httpClient(ctx, request.ClusterID, "s3")
	if err != nil {
		return cephdomain.ActionResult{}, err
	}
	if s.topicOwners == nil || credential.SessionToken != "" {
		return cephdomain.ActionResult{}, failure("capability_unavailable", "topic creation requires native owner verification and permanent S3 credentials", false)
	}
	scope, owner, err := s.topicOwners.VerifyTopicOwner(ctx, request.ClusterID, text("owner_uid"), credential.AccessKey, credential.SecretKey)
	if err != nil || scope != strings.SplitN(arn, ":", 6)[4] || owner == "" {
		return cephdomain.ActionResult{}, failure("pre_check_failed", "configured S3 key does not verify the requested topic owner and scope", false)
	}
	api, err := s3.New(endpoint.URL, s3.Credentials{AccessKey: credential.AccessKey, SecretKey: credential.SecretKey, Region: credential.Region}, client)
	if err != nil {
		return cephdomain.ActionResult{}, failure("invalid_credential", "SNS credentials or endpoint invalid", false)
	}
	if err := api.RequireTopicAbsent(ctx, arn); err != nil {
		return cephdomain.ActionResult{}, failure("pre_check_failed", "target topic exists or absence could not be verified; no creation submitted", false)
	}
	returned, err := api.CreateTopic(ctx, input)
	if err != nil || returned != arn {
		return cephdomain.ActionResult{}, failure("sns_failed", "topic creation outcome or returned ARN could not be verified; inspect native state before retrying", false)
	}
	attrs, err := api.GetTopicAttributes(ctx, arn)
	if err != nil || !createdTopicMatches(attrs, input, arn, owner) {
		return cephdomain.ActionResult{}, failure("post_check_failed", "topic creation was submitted but complete configuration could not be verified; topic or queue may already exist", false)
	}
	return cephdomain.ActionResult{}, nil
}

func createdTopicMatches(attrs map[string]string, p s3.TopicCreate, arn, owner string) bool {
	if len(attrs) != 6 || attrs["TopicArn"] != arn || attrs["Name"] != p.Name || attrs["User"] != owner || attrs["Policy"] != p.Policy || attrs["OpaqueData"] != p.OpaqueData {
		return false
	}
	dest, err := topicDestination(attrs)
	if err != nil || len(dest) != 8 {
		return false
	}
	secret, _ := s3.ValidateTopicEndpoint(p.Endpoint)
	expected := map[string]any{"EndpointAddress": p.Endpoint, "EndpointArgs": p.EndpointArgs(), "EndpointTopic": p.Name, "HasStoredSecret": secret, "Persistent": p.Persistent, "TimeToLive": p.TTL, "MaxRetries": p.MaxRetries, "RetrySleepDuration": p.RetrySleep}
	for key, want := range expected {
		var actual any
		if json.Unmarshal(dest[key], &actual) != nil || actual != want {
			return false
		}
	}
	return true
}
