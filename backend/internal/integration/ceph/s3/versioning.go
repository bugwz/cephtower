package s3

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// BucketVersioningStatus returns empty for the native never-versioned response.
// An empty configuration is not equivalent to Suspended.
func BucketVersioningStatus(body []byte) (string, error) {
	configuration, err := BucketVersioning(body)
	return configuration.Status, err
}

type BucketVersioningConfiguration struct {
	Status    string  `json:"status"`
	MFADelete *string `json:"mfa_delete"`
}

func BucketVersioning(body []byte) (BucketVersioningConfiguration, error) {
	result := BucketVersioningConfiguration{}
	if err := ValidateBucketConfiguration("versioning", body); err != nil {
		return result, err
	}
	var document struct {
		Status  []tagText  `xml:"Status"`
		MFA     []tagText  `xml:"MfaDelete"`
		Unknown []xml.Name `xml:",any"`
		Text    string     `xml:",chardata"`
	}
	invalid := fmt.Errorf("invalid native bucket versioning response")
	if xml.Unmarshal(body, &document) != nil || len(document.Status) > 1 || len(document.MFA) > 1 || len(document.Unknown) > 0 || strings.TrimSpace(document.Text) != "" {
		return result, invalid
	}
	if len(document.MFA) == 1 {
		mfa := document.MFA[0]
		if len(mfa.Unknown) > 0 || (mfa.Text != "Enabled" && mfa.Text != "Disabled") || len(document.Status) == 0 {
			return result, invalid
		}
		result.MFADelete = &document.MFA[0].Text
	}
	if len(document.Status) == 0 {
		return result, nil
	}
	status := document.Status[0]
	if len(status.Unknown) > 0 || (status.Text != "Enabled" && status.Text != "Suspended") {
		return result, invalid
	}
	result.Status = status.Text
	return result, nil
}
