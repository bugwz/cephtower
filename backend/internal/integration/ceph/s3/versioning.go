package s3

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// BucketVersioningStatus returns empty for the native never-versioned response.
// An empty configuration is not equivalent to Suspended.
func BucketVersioningStatus(body []byte) (string, error) {
	if err := ValidateBucketConfiguration("versioning", body); err != nil {
		return "", err
	}
	var document struct {
		Status  []tagText  `xml:"Status"`
		MFA     []tagText  `xml:"MfaDelete"`
		Unknown []xml.Name `xml:",any"`
		Text    string     `xml:",chardata"`
	}
	invalid := fmt.Errorf("invalid native bucket versioning response")
	if xml.Unmarshal(body, &document) != nil || len(document.Status) > 1 || len(document.MFA) > 1 || len(document.Unknown) > 0 || strings.TrimSpace(document.Text) != "" {
		return "", invalid
	}
	if len(document.MFA) == 1 {
		mfa := document.MFA[0]
		if len(mfa.Unknown) > 0 || (mfa.Text != "Enabled" && mfa.Text != "Disabled") || len(document.Status) == 0 {
			return "", invalid
		}
	}
	if len(document.Status) == 0 {
		return "", nil
	}
	status := document.Status[0]
	if len(status.Unknown) > 0 || (status.Text != "Enabled" && status.Text != "Suspended") {
		return "", invalid
	}
	return status.Text, nil
}
