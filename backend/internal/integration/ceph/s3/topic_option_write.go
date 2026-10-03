package s3

import (
	"context"
	"fmt"
	"strings"
)

func ValidateTopicOption(name, value string) error {
	switch name {
	case "verify-ssl", "use-ssl", "cloudevents", "ca-location", "amqp-exchange", "amqp-ack-level", "kafka-ack-level", "mechanism":
		if safeTopicOption(name, value) {
			return nil
		}
	}
	return fmt.Errorf("unsupported topic option or unsafe value")
}

// Model the native substring replacement, rejecting any match that is not the
// exact selected key. The raw secret-bearing argument string stays in memory.
func PrepareTopicOption(raw, name, expectedStatus, expected, value string) (string, error) {
	invalid := fmt.Errorf("topic option snapshot or native replacement is unsafe")
	if ValidateTopicOption(name, value) != nil {
		return "", invalid
	}
	options, valid := TopicOptions(raw)
	if !valid {
		return "", invalid
	}
	matched := false
	for _, option := range options {
		if option.Name == name {
			if option.Status != expectedStatus {
				return "", invalid
			}
			switch option.Status {
			case "unset":
				if expected != "" {
					return "", invalid
				}
			case "returned":
				if option.Value == nil || *option.Value != expected || expected == value {
					return "", invalid
				}
			default:
				return "", invalid
			}
			matched = true
		}
	}
	if !matched {
		return "", invalid
	}
	replacement := name + "=" + value
	pos := strings.Index(raw, name)
	if pos < 0 {
		if expectedStatus != "unset" {
			return "", invalid
		}
		return raw + "&" + replacement, nil
	}
	if expectedStatus != "returned" || (pos > 0 && raw[pos-1] != '&') || !strings.HasPrefix(raw[pos:], name+"=") {
		return "", invalid
	}
	end := strings.Index(raw[pos:], "&")
	if end < 0 {
		end = len(raw)
	} else {
		end += pos
	}
	return raw[:pos] + replacement + raw[end:], nil
}

func (c *Client) SetTopicOption(ctx context.Context, arn, name, value string) error {
	if err := ValidateTopicOption(name, value); err != nil {
		return err
	}
	return c.setTopicAttribute(ctx, arn, name, value)
}
