package s3

import (
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

// BucketEncryptionConfiguration reflects RGWBucketEncryptionConfig, including
// the distinction between a stored empty configuration and a stored empty rule.
type BucketEncryptionConfiguration struct {
	RuleExists       bool   `json:"rule_exists"`
	Algorithm        string `json:"algorithm"`
	KMSMasterKeyID   string `json:"kms_master_key_id"`
	BucketKeyEnabled bool   `json:"bucket_key_enabled"`
}

func BucketEncryption(body []byte) (BucketEncryptionConfiguration, error) {
	if err := ValidateBucketConfiguration("encryption", body); err != nil {
		return BucketEncryptionConfiguration{}, err
	}
	return parseBucketEncryption(body)
}

func parseBucketEncryption(body []byte) (BucketEncryptionConfiguration, error) {
	var document struct {
		Rules []struct {
			Defaults []struct {
				Algorithms []tagText  `xml:"SSEAlgorithm"`
				Keys       []tagText  `xml:"KMSMasterKeyID"`
				Text       string     `xml:",chardata"`
				Unknown    []xml.Name `xml:",any"`
			} `xml:"ApplyServerSideEncryptionByDefault"`
			Enabled []tagText  `xml:"BucketKeyEnabled"`
			Text    string     `xml:",chardata"`
			Unknown []xml.Name `xml:",any"`
		} `xml:"Rule"`
		Text    string     `xml:",chardata"`
		Unknown []xml.Name `xml:",any"`
	}
	var result BucketEncryptionConfiguration
	invalid := func() (BucketEncryptionConfiguration, error) {
		return BucketEncryptionConfiguration{}, fmt.Errorf("encryption requires at most one Rule, one default encryption block and unique scalar fields without unknown elements")
	}
	if xml.Unmarshal(body, &document) != nil || len(document.Rules) > 1 || len(document.Unknown) != 0 || strings.TrimSpace(document.Text) != "" {
		return invalid()
	}
	if len(document.Rules) == 0 {
		return result, nil
	}
	result.RuleExists = true
	rule := document.Rules[0]
	if len(rule.Defaults) > 1 || len(rule.Enabled) > 1 || len(rule.Unknown) != 0 || strings.TrimSpace(rule.Text) != "" {
		return invalid()
	}
	if len(rule.Defaults) == 1 {
		defaults := rule.Defaults[0]
		if len(defaults.Algorithms) > 1 || len(defaults.Keys) > 1 || len(defaults.Unknown) != 0 || strings.TrimSpace(defaults.Text) != "" {
			return invalid()
		}
		if len(defaults.Algorithms) == 1 {
			if len(defaults.Algorithms[0].Unknown) != 0 {
				return invalid()
			}
			result.Algorithm = defaults.Algorithms[0].Text
		}
		if len(defaults.Keys) == 1 {
			if len(defaults.Keys[0].Unknown) != 0 {
				return invalid()
			}
			result.KMSMasterKeyID = defaults.Keys[0].Text
		}
	}
	if len(rule.Enabled) == 1 {
		if len(rule.Enabled[0].Unknown) != 0 {
			return invalid()
		}
		value := rule.Enabled[0].Text
		switch {
		case strings.EqualFold(value, "true"):
			result.BucketKeyEnabled = true
		case strings.EqualFold(value, "false"):
		default:
			// RGW's XML bool decoder accepts signed 32-bit decimal integers,
			// with C whitespace, and converts every nonzero value to true.
			number, err := strconv.ParseInt(strings.Trim(value, " \t\r\n\v\f"), 10, 32)
			if err != nil {
				return invalid()
			}
			result.BucketKeyEnabled = number != 0
		}
	}
	return result, nil
}
