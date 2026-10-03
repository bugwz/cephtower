package s3

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// Replication summaries supplement, rather than replace, the complete XML.
// Filters, zone extensions and optional destination settings remain in document.
type BucketReplicationConfiguration struct {
	Role  string                  `json:"role"`
	Rules []BucketReplicationRule `json:"rules"`
}
type BucketReplicationRule struct {
	ID                string  `json:"id"`
	Status            string  `json:"status"`
	Priority          *string `json:"priority"`
	DestinationBucket string  `json:"destination_bucket"`
}

func BucketReplication(body []byte) (BucketReplicationConfiguration, error) {
	result := BucketReplicationConfiguration{Rules: []BucketReplicationRule{}}
	if err := validateConfigurationXML(body, "ReplicationConfiguration"); err != nil {
		return result, err
	}
	var root lifecycleField
	if err := xml.Unmarshal(body, &root); err != nil {
		return result, err
	}
	invalid := fmt.Errorf("invalid native ReplicationConfiguration structure")
	if strings.TrimSpace(root.Text) != "" {
		return result, invalid
	}
	roleSeen := false
	for _, node := range root.Children {
		switch node.XMLName.Local {
		case "Role":
			if roleSeen || len(node.Children) != 0 {
				return result, invalid
			}
			roleSeen, result.Role = true, node.Text
		case "Rule":
			if strings.TrimSpace(node.Text) != "" {
				return result, invalid
			}
			rule := BucketReplicationRule{}
			seen := map[string]bool{}
			for _, field := range node.Children {
				name := field.XMLName.Local
				// All standard top-level rule elements are singletons. Unknown extensions
				// stay in the raw document, but must not make the summary ambiguous.
				if seen[name] {
					return result, invalid
				}
				seen[name] = true
				switch name {
				case "ID", "Status", "Priority":
					if len(field.Children) != 0 {
						return result, invalid
					}
					switch name {
					case "ID":
						rule.ID = field.Text
					case "Status":
						rule.Status = field.Text
					case "Priority":
						value := field.Text
						rule.Priority = &value
					}
				case "Destination":
					if strings.TrimSpace(field.Text) != "" {
						return result, invalid
					}
					bucketSeen := false
					for _, destination := range field.Children {
						if destination.XMLName.Local == "Bucket" {
							if bucketSeen || len(destination.Children) != 0 {
								return result, invalid
							}
							bucketSeen, rule.DestinationBucket = true, destination.Text
						}
					}
					if !bucketSeen {
						return result, invalid
					}
				}
			}
			if !seen["Status"] || rule.Status == "" || !seen["Destination"] {
				return result, invalid
			}
			result.Rules = append(result.Rules, rule)
		default:
			return result, invalid
		}
	}
	return result, nil
}
