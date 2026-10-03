package s3

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
)

// Match the current RGW parse_url_userinfo grammar before sending secrets.
var topicEndpointPattern = regexp.MustCompile(`^[[:alpha:]]+://(([^:\s]+):([^@\s]+)@)?([[:alnum:].:-]+)(/[[:print:]]*)?$`)

func ValidateTopicEndpoint(value string) (bool, error) {
	if value == "" {
		return false, nil
	}
	parsed, err := url.Parse(value)
	if err != nil || !topicEndpointPattern.MatchString(value) || parsed.Hostname() == "" {
		return false, fmt.Errorf("invalid topic push endpoint")
	}
	switch parsed.Scheme {
	case "http", "https", "amqp", "amqps", "kafka":
	default:
		return false, fmt.Errorf("unsupported topic push endpoint protocol")
	}
	if parsed.User != nil {
		password, present := parsed.User.Password()
		if parsed.User.Username() == "" || !present || password == "" {
			return false, fmt.Errorf("topic endpoint credentials must be a complete pair")
		}
		return true, nil
	}
	return false, nil
}

// A shared projection prevents API mutation checks from disagreeing with CLI
// inventory redaction. It is not a fingerprint of hidden credentials or args.
func RedactTopicEndpoint(value string) (string, bool, bool) {
	if value == "" {
		return "", false, true
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Scheme == "" || parsed.Hostname() == "" || parsed.OmitHost {
		return "", true, false
	}
	redacted := parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != ""
	parsed.User, parsed.RawQuery, parsed.Fragment, parsed.RawFragment, parsed.ForceQuery = nil, "", "", "", false
	return parsed.String(), redacted, true
}

func (c *Client) SetTopicEndpoint(ctx context.Context, arn, value string) error {
	if _, err := ValidateTopicEndpoint(value); err != nil {
		return err
	}
	return c.setTopicAttribute(ctx, arn, "push-endpoint", value)
}
