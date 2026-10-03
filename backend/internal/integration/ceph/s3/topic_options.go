package s3

import (
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type TopicOption struct {
	Name   string  `json:"name"`
	Status string  `json:"status"`
	Value  *string `json:"value,omitempty"`
}

var topicOptionNames = []string{"verify-ssl", "use-ssl", "cloudevents", "ca-location", "amqp-version", "amqp-exchange", "amqp-ack-level", "http-ack-level", "kafka-ack-level", "mechanism", "kafka-brokers"}
var topicPathValue = regexp.MustCompile(`^[A-Za-z0-9_./ -]+$`)
var topicNameValue = regexp.MustCompile(`^[A-Za-z0-9_.-]{0,256}$`)
var topicBrokerValue = regexp.MustCompile(`^[A-Za-z0-9.-]+(:[0-9]+)?(,[A-Za-z0-9.-]+(:[0-9]+)?)*$`)

// Only known non-credential parameters enter inventory. RGW decodes each entire
// '&'-separated segment BEFORE splitting name/value, unlike url.ParseQuery.
func TopicOptions(raw string) ([]TopicOption, bool) {
	values := map[string][]string{}
	if raw != "" {
		for _, segment := range strings.Split(strings.TrimPrefix(raw, "?"), "&") {
			decoded, err := url.QueryUnescape(segment)
			if err != nil {
				return nil, false
			}
			name, value, _ := strings.Cut(decoded, "=")
			values[name] = append(values[name], value)
		}
	}
	result := make([]TopicOption, 0, len(topicOptionNames))
	for _, name := range topicOptionNames {
		option := TopicOption{Name: name, Status: "unset"}
		if items := values[name]; len(items) > 1 {
			option.Status = "duplicate"
		} else if len(items) == 1 {
			option.Status = "hidden_invalid"
			if safeTopicOption(name, items[0]) {
				value := items[0]
				option.Status = "returned"
				option.Value = &value
			}
		}
		result = append(result, option)
	}
	return result, true
}

func safeTopicOption(name, value string) bool {
	switch name {
	case "verify-ssl", "use-ssl", "cloudevents":
		return strings.EqualFold(value, "true") || strings.EqualFold(value, "false")
	case "ca-location":
		return value == "" || len(value) <= 4096 && topicPathValue.MatchString(value)
	case "amqp-version":
		return value == "0-9-1" || value == "1-0"
	case "amqp-exchange":
		return topicNameValue.MatchString(value)
	case "amqp-ack-level":
		return value == "none" || value == "broker" || value == "routable"
	case "kafka-ack-level":
		return value == "none" || value == "broker"
	case "http-ack-level":
		if value == "any" || value == "non-error" {
			return true
		}
		n, err := strconv.Atoi(value)
		return err == nil && n >= 100 && n < 600 && strconv.Itoa(n) == value
	case "mechanism":
		return value == "PLAIN" || value == "SCRAM-SHA-256" || value == "SCRAM-SHA-512" || value == "GSSAPI" || value == "OAUTHBEARER"
	case "kafka-brokers":
		return len(value) <= 4096 && topicBrokerValue.MatchString(value)
	}
	return false
}
