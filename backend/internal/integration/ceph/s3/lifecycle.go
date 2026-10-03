package s3

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// Validate the rule envelope without reducing native action/filter subtrees to
// a smaller supported subset. Their detailed semantics remain RGW's authority.
func validateBucketLifecycle(body []byte) error {
	var document struct {
		Rules []struct {
			Fields []struct {
				XMLName  xml.Name
				Text     string     `xml:",chardata"`
				Children []xml.Name `xml:",any"`
			} `xml:",any"`
			Text string `xml:",chardata"`
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
