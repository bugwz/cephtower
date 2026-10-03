package s3

import (
	"encoding/xml"
	"fmt"
	"strconv"
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
		if len(rule.Ages) == 1 {
			if _, err := corsMaxAge(rule.Ages[0].Text); err != nil {
				return err
			}
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

// Match RGW's strtoull(base 10), end-pointer check, uint32 conversion and
// CORS_MAX_AGE_INVALID sentinel. Nil means the native XML omits this field.
func corsMaxAge(text string) (*uint32, error) {
	invalid := fmt.Errorf("CORS MaxAgeSeconds must use the native decimal integer syntax")
	if text == "" {
		zero := uint32(0)
		return &zero, nil
	}
	value := strings.TrimLeft(text, " \t\r\n\v\f")
	negative := strings.HasPrefix(value, "-")
	if strings.HasPrefix(value, "+") || negative {
		value = value[1:]
	}
	if value == "" {
		return nil, invalid
	}
	for _, digit := range value {
		if digit < '0' || digit > '9' {
			return nil, invalid
		}
	}
	number, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		// Syntax was checked above. strtoull returns ULLONG_MAX on overflow;
		// RGW does not inspect errno and treats it as an absent maximum age.
		return nil, nil
	}
	if negative {
		number = -number
	}
	if number >= 0xffffffff {
		return nil, nil
	}
	seconds := uint32(number)
	return &seconds, nil
}
