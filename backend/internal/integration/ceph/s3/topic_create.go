package s3

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"sort"
	"strings"
)

type TopicCreate struct {
	Name, Endpoint, OpaqueData, Policy string
	Persistent                         bool
	TTL, MaxRetries, RetrySleep        string
	Options                            map[string]string
}

var topicCreateName = regexp.MustCompile(`^[A-Za-z0-9_-]{1,256}$`)

func (p TopicCreate) Validate() error {
	invalid := fmt.Errorf("invalid topic creation configuration")
	if !topicCreateName.MatchString(p.Name) {
		return invalid
	}
	if _, err := ValidateTopicEndpoint(p.Endpoint); err != nil {
		return invalid
	}
	if p.Policy != "" && ValidateBucketConfiguration("policy", []byte(p.Policy)) != nil {
		return invalid
	}
	for key, value := range map[string]string{"time_to_live": p.TTL, "max_retries": p.MaxRetries, "retry_sleep_duration": p.RetrySleep} {
		if _, err := TopicAttributeWireValue(key, value); err != nil {
			return invalid
		}
	}
	for key, value := range p.Options {
		if !safeTopicOption(key, value) {
			return invalid
		}
	}
	return nil
}
func (p TopicCreate) EndpointArgs() string {
	values := map[string]string{"Version": "2010-03-31"}
	for k, v := range p.Options {
		values[k] = v
	}
	keys := make([]string, 0, len(values))
	for k := range values {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, k+"="+values[k])
	}
	return strings.Join(parts, "&")
}

func (c *Client) RequireTopicAbsent(ctx context.Context, arn string) error {
	_, err := c.topicRawRequest(ctx, url.Values{"Action": {"GetTopicAttributes"}, "TopicArn": {arn}})
	var response *ResponseError
	if errors.As(err, &response) && response.StatusCode == http.StatusNotFound && validateConfigurationXML([]byte(response.message), "ErrorResponse") == nil {
		var root lifecycleField
		if xml.Unmarshal([]byte(response.message), &root) == nil && len(root.Children) == 1 && root.Children[0].XMLName.Local == "Error" {
			count := 0
			code := ""
			for _, field := range root.Children[0].Children {
				if field.XMLName.Local == "Code" {
					count++
					if len(field.Children) != 0 {
						return fmt.Errorf("topic absence could not be verified")
					}
					code = field.Text
				}
			}
			if count == 1 && code == "NotFound" {
				return nil
			}
		}
	}
	return fmt.Errorf("topic exists or absence could not be verified")
}
func (c *Client) CreateTopic(ctx context.Context, p TopicCreate) (string, error) {
	if err := p.Validate(); err != nil {
		return "", err
	}
	values := url.Values{"Action": {"CreateTopic"}, "Name": {p.Name}, "push-endpoint": {p.Endpoint}, "OpaqueData": {p.OpaqueData}, "Policy": {p.Policy}, "persistent": {"false"}}
	if p.Persistent {
		values.Set("persistent", "true")
	}
	for key, value := range map[string]string{"time_to_live": p.TTL, "max_retries": p.MaxRetries, "retry_sleep_duration": p.RetrySleep} {
		wire, _ := TopicAttributeWireValue(key, value)
		values.Set(key, wire)
	}
	for k, v := range p.Options {
		values.Set(k, v)
	}
	body, err := c.topicRequest(ctx, values)
	if err != nil {
		return "", err
	}
	invalid := fmt.Errorf("invalid SNS creation response")
	if validateConfigurationXML(body, "CreateTopicResponse") != nil {
		return "", invalid
	}
	var root lifecycleField
	if xml.Unmarshal(body, &root) != nil {
		return "", invalid
	}
	count := 0
	arn := ""
	for _, result := range root.Children {
		if result.XMLName.Local == "CreateTopicResult" {
			count++
			if len(result.Children) != 1 || result.Children[0].XMLName.Local != "TopicArn" || len(result.Children[0].Children) != 0 {
				return "", invalid
			}
			arn = result.Children[0].Text
		}
	}
	if count != 1 || arn == "" {
		return "", invalid
	}
	return arn, nil
}
