package s3

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// SNS shares the RGW endpoint and credentials, but uses its own signing scope.
// Use a form body rather than putting policy contents in the URL or error text.
func (c *Client) topicRequest(ctx context.Context, values url.Values) ([]byte, error) {
	if c.base.Scheme != "https" {
		return nil, fmt.Errorf("SNS topic operations require HTTPS")
	}
	values.Set("Version", "2010-03-31")
	body, _, err := c.requestSignedTarget(ctx, http.MethodPost, "", nil, []byte(values.Encode()), http.Header{"Content-Type": {"application/x-www-form-urlencoded"}}, "sns")
	if err != nil {
		return nil, fmt.Errorf("SNS request failed")
	}
	return body, nil
}

func (c *Client) GetTopicAttributes(ctx context.Context, arn string) (map[string]string, error) {
	body, err := c.topicRequest(ctx, url.Values{"Action": {"GetTopicAttributes"}, "TopicArn": {arn}})
	if err != nil {
		return nil, err
	}
	return parseTopicAttributes(body)
}

func (c *Client) SetTopicPolicy(ctx context.Context, arn, policy string) error {
	if policy != "" {
		if err := ValidateBucketConfiguration("policy", []byte(policy)); err != nil {
			return err
		}
	}
	return c.setTopicAttribute(ctx, arn, "Policy", policy)
}

// Numeric attributes are parsed by RGWHTTPArgs::get_int, not an unsigned parser.
// None is the displayed global-default sentinel; send its explicit signed value.
func TopicAttributeWireValue(attribute, value string) (string, error) {
	switch attribute {
	case "OpaqueData":
		return value, nil
	case "persistent":
		if value == "true" || value == "false" {
			return value, nil
		}
	case "time_to_live", "max_retries", "retry_sleep_duration":
		if value == "None" {
			return "-1", nil
		}
		if n, err := strconv.ParseUint(value, 10, 31); err == nil && strconv.FormatUint(n, 10) == value {
			return value, nil
		}
	}
	return "", fmt.Errorf("invalid SNS topic attribute or value")
}

func (c *Client) SetTopicAttribute(ctx context.Context, arn, attribute, value string) error {
	wire, err := TopicAttributeWireValue(attribute, value)
	if err != nil {
		return err
	}
	return c.setTopicAttribute(ctx, arn, attribute, wire)
}

func (c *Client) setTopicAttribute(ctx context.Context, arn, attribute, value string) error {
	body, err := c.topicRequest(ctx, url.Values{"Action": {"SetTopicAttributes"}, "TopicArn": {arn}, "AttributeName": {attribute}, "AttributeValue": {value}})
	if err != nil {
		return err
	}
	return validateConfigurationXML(body, "SetTopicAttributesResponse")
}

func parseTopicAttributes(body []byte) (map[string]string, error) {
	if err := validateConfigurationXML(body, "GetTopicAttributesResponse"); err != nil {
		return nil, err
	}
	var root lifecycleField
	if err := xml.Unmarshal(body, &root); err != nil {
		return nil, err
	}
	invalid := fmt.Errorf("invalid SNS topic attributes response")
	// Native SNS entries are single key/value pairs, including empty values.
	one := func(node lifecycleField, name string) (lifecycleField, bool) {
		var found lifecycleField
		count := 0
		for _, child := range node.Children {
			if child.XMLName.Local == name {
				found = child
				count++
			}
		}
		return found, count == 1
	}
	result, ok := one(root, "GetTopicAttributesResult")
	if !ok {
		return nil, invalid
	}
	attributes, ok := one(result, "Attributes")
	if !ok || len(result.Children) != 1 || strings.TrimSpace(result.Text) != "" || strings.TrimSpace(attributes.Text) != "" {
		return nil, invalid
	}
	values := map[string]string{}
	for _, entry := range attributes.Children {
		if entry.XMLName.Local != "entry" || len(entry.Children) != 2 || strings.TrimSpace(entry.Text) != "" {
			return nil, invalid
		}
		key, keyOK := one(entry, "key")
		value, valueOK := one(entry, "value")
		if !keyOK || !valueOK || len(key.Children) != 0 || len(value.Children) != 0 || key.Text == "" {
			return nil, invalid
		}
		if _, exists := values[key.Text]; exists {
			return nil, invalid
		}
		values[key.Text] = value.Text
	}
	for _, key := range []string{"User", "Name", "EndPoint", "TopicArn", "OpaqueData", "Policy"} {
		if _, ok := values[key]; !ok {
			return nil, invalid
		}
	}
	return values, nil
}
