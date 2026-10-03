package s3

import (
	"encoding/xml"
	"fmt"
	"strings"
)

func validateBucketCORS(body []byte) error {
	var document struct {
		Rules []struct {
			IDs     []tagText  `xml:"ID"`
			Origins []tagText  `xml:"AllowedOrigin"`
			Methods []tagText  `xml:"AllowedMethod"`
			Headers []tagText  `xml:"AllowedHeader"`
			Exposed []tagText  `xml:"ExposeHeader"`
			Ages    []tagText  `xml:"MaxAgeSeconds"`
			Unknown []xml.Name `xml:",any"`
			Text    string     `xml:",chardata"`
		} `xml:"CORSRule"`
		Unknown []xml.Name `xml:",any"`
		Text    string     `xml:",chardata"`
	}
	invalid := fmt.Errorf("CORS requires rules with nonempty origins, supported methods, IDs up to 255 UTF-8 bytes and at most one wildcard per origin or allowed header")
	if xml.Unmarshal(body, &document) != nil || len(document.Rules) == 0 || len(document.Unknown) != 0 || strings.TrimSpace(document.Text) != "" {
		return invalid
	}
	for _, rule := range document.Rules {
		if len(rule.Origins) == 0 || len(rule.IDs) > 1 || len(rule.Ages) > 1 || len(rule.Unknown) > 0 || strings.TrimSpace(rule.Text) != "" {
			return invalid
		}
		for _, fields := range [][]tagText{rule.IDs, rule.Origins, rule.Methods, rule.Headers, rule.Exposed, rule.Ages} {
			for _, field := range fields {
				if len(field.Unknown) > 0 {
					return invalid
				}
			}
		}
		if len(rule.IDs) == 1 && len(rule.IDs[0].Text) > 255 {
			return invalid
		}
		for _, fields := range [][]tagText{rule.Origins, rule.Headers} {
			for _, field := range fields {
				if field.Text == "" || strings.Count(field.Text, "*") > 1 {
					return invalid
				}
			}
		}
		for _, method := range rule.Methods {
			switch strings.ToUpper(method.Text) {
			case "GET", "PUT", "DELETE", "HEAD", "POST", "COPY":
			default:
				return invalid
			}
		}
	}
	return nil
}
