package s3

import (
	"encoding/xml"
	"fmt"
	"strings"
)

type lifecycleField struct {
	XMLName  xml.Name
	Text     string           `xml:",chardata"`
	Children []lifecycleField `xml:",any"`
}

// Validate rule and action structure; numeric/date and filter semantics are
// still checked by RGW rather than replaced with a reduced feature subset.
func validateBucketLifecycle(body []byte) error {
	var document struct {
		Rules []struct {
			Fields []lifecycleField `xml:",any"`
			Text   string           `xml:",chardata"`
		} `xml:"Rule"`
		Unknown []xml.Name `xml:",any"`
		Text    string     `xml:",chardata"`
	}
	invalid := fmt.Errorf("lifecycle requires Rule entries with one Enabled/Disabled Status, a Filter or Prefix, and at least one expiration or transition action")
	if xml.Unmarshal(body, &document) != nil || len(document.Rules) == 0 || len(document.Unknown) > 0 || strings.TrimSpace(document.Text) != "" {
		return invalid
	}
	for _, rule := range document.Rules {
		if strings.TrimSpace(rule.Text) != "" {
			return invalid
		}
		counts := map[string]int{}
		actions := 0
		for _, field := range rule.Fields {
			name := field.XMLName.Local
			counts[name]++
			switch name {
			case "ID", "Prefix", "Status":
				if len(field.Children) > 0 {
					return invalid
				}
				if name == "Status" && field.Text != "Enabled" && field.Text != "Disabled" {
					return invalid
				}
			case "Filter":
				if strings.TrimSpace(field.Text) != "" {
					return invalid
				}
			case "Expiration", "NoncurrentVersionExpiration", "AbortIncompleteMultipartUpload", "Transition", "NoncurrentVersionTransition":
				if strings.TrimSpace(field.Text) != "" || len(field.Children) == 0 {
					return invalid
				}
				if err := validateLifecycleAction(field); err != nil {
					return err
				}
				actions++
			default:
				return invalid
			}
			if name != "Transition" && name != "NoncurrentVersionTransition" && counts[name] > 1 {
				return invalid
			}
		}
		if counts["Status"] != 1 || counts["Filter"]+counts["Prefix"] != 1 || actions == 0 {
			return invalid
		}
	}
	return nil
}

func validateLifecycleAction(action lifecycleField) error {
	name := action.XMLName.Local
	allowed := map[string][]string{
		"Expiration":                     {"Days", "Date", "ExpiredObjectDeleteMarker"},
		"NoncurrentVersionExpiration":    {"NoncurrentDays", "NewerNoncurrentVersions"},
		"AbortIncompleteMultipartUpload": {"DaysAfterInitiation"},
		"Transition":                     {"Days", "Date", "StorageClass"},
		"NoncurrentVersionTransition":    {"NoncurrentDays", "StorageClass"},
	}[name]
	counts := map[string]int{}
	invalid := fmt.Errorf("invalid lifecycle %s fields: check required, exclusive and duplicate scalar elements", name)
	for _, child := range action.Children {
		key := child.XMLName.Local
		known := false
		for _, field := range allowed {
			if key == field {
				known = true
				break
			}
		}
		if !known || len(child.Children) > 0 {
			return invalid
		}
		counts[key]++
		if counts[key] > 1 {
			return invalid
		}
	}
	switch name {
	case "Expiration":
		if counts["Days"]+counts["Date"]+counts["ExpiredObjectDeleteMarker"] != 1 {
			return invalid
		}
	case "Transition":
		if counts["Days"]+counts["Date"] != 1 || counts["StorageClass"] != 1 {
			return invalid
		}
	case "NoncurrentVersionTransition":
		if counts["NoncurrentDays"] != 1 || counts["StorageClass"] != 1 {
			return invalid
		}
	case "NoncurrentVersionExpiration":
		if counts["NoncurrentDays"] != 1 {
			return invalid
		}
	case "AbortIncompleteMultipartUpload":
		if counts["DaysAfterInitiation"] != 1 {
			return invalid
		}
	}
	return nil
}
