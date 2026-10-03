package s3

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type BucketNotificationFilter struct {
	Kind  string `json:"kind"`
	Name  string `json:"name"`
	Value string `json:"value"`
}

type BucketNotification struct {
	ID      string                     `json:"id"`
	Topic   string                     `json:"topic"`
	Events  []string                   `json:"events"`
	Filters []BucketNotificationFilter `json:"filters"`
}

// BucketNotifications reads the native S3 projection, not topic metadata (which
// also contains delivery credentials). Preserve order, duplicate IDs and unknown
// event/filter names as returned; this function does not validate a PUT request.
func BucketNotifications(body []byte) ([]BucketNotification, error) {
	if err := validateConfigurationXML(body, "NotificationConfiguration"); err != nil {
		return nil, err
	}
	var root lifecycleField
	if err := xml.Unmarshal(body, &root); err != nil {
		return nil, fmt.Errorf("invalid notification XML")
	}
	invalid := fmt.Errorf("invalid native NotificationConfiguration structure")
	container := func(node lifecycleField) bool { return strings.TrimSpace(node.Text) == "" }
	if !container(root) {
		return nil, invalid
	}
	result := []BucketNotification{}
	for _, node := range root.Children {
		if node.XMLName.Local != "TopicConfiguration" || !container(node) {
			return nil, invalid
		}
		rule := BucketNotification{Events: []string{}, Filters: []BucketNotificationFilter{}}
		seen := map[string]bool{}
		for _, field := range node.Children {
			name := field.XMLName.Local
			if name != "Event" && seen[name] {
				return nil, invalid
			}
			seen[name] = true
			switch name {
			case "Id", "Topic", "Event":
				if len(field.Children) != 0 {
					return nil, invalid
				}
				switch name {
				case "Id":
					rule.ID = field.Text
				case "Topic":
					rule.Topic = field.Text
				case "Event":
					rule.Events = append(rule.Events, field.Text)
				}
			case "Filter":
				if !container(field) {
					return nil, invalid
				}
				kinds := map[string]bool{}
				for _, group := range field.Children {
					kind := group.XMLName.Local
					if kinds[kind] || !container(group) || (kind != "S3Key" && kind != "S3Metadata" && kind != "S3Tags") {
						return nil, invalid
					}
					kinds[kind] = true
					for _, filter := range group.Children {
						if filter.XMLName.Local != "FilterRule" || !container(filter) || len(filter.Children) != 2 {
							return nil, invalid
						}
						entry := BucketNotificationFilter{Kind: kind}
						fields := map[string]bool{}
						for _, pair := range filter.Children {
							key := pair.XMLName.Local
							if fields[key] || len(pair.Children) != 0 || (key != "Name" && key != "Value") {
								return nil, invalid
							}
							fields[key] = true
							if key == "Name" {
								entry.Name = pair.Text
							} else {
								entry.Value = pair.Text
							}
						}
						rule.Filters = append(rule.Filters, entry)
					}
				}
			default:
				return nil, invalid
			}
		}
		if !seen["Id"] || !seen["Topic"] {
			return nil, invalid
		}
		result = append(result, rule)
	}
	return result, nil
}
