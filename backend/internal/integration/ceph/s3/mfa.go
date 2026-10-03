package s3

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
)

var mfaSerial = regexp.MustCompile(`^[!-~]+$`)
var mfaPIN = regexp.MustCompile(`^[0-9]{1,128}$`)

func ValidateBucketMFA(status, mfa, serial, token string) error {
	if (status != "Enabled" && status != "Suspended") || (mfa != "Enabled" && mfa != "Disabled") || len(serial) > 1024 || !mfaSerial.MatchString(serial) || !mfaPIN.MatchString(token) {
		return fmt.Errorf("valid versioning/MFA states, a non-whitespace ASCII device serial and a numeric code are required")
	}
	return nil
}

func (c *Client) PutBucketMFA(ctx context.Context, bucket, status, mfa, serial, token string) error {
	if err := ValidateBucketMFA(status, mfa, serial, token); err != nil {
		return err
	}
	if c.base.Scheme != "https" {
		return fmt.Errorf("MFA operations require HTTPS")
	}
	// Never replay the one-time credential to a redirect target, including HTTP
	// downgrade redirects. Copy the client rather than mutate shared transport state.
	requestClient := *c
	httpClient := *c.http
	httpClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	requestClient.http = &httpClient
	body := []byte(`<VersioningConfiguration><Status>` + status + `</Status><MfaDelete>` + mfa + `</MfaDelete></VersioningConfiguration>`)
	_, _, err := requestClient.requestWithHeaders(ctx, http.MethodPut, bucket, url.Values{"versioning": {""}}, body, http.Header{"X-Amz-Mfa": {serial + " " + token}})
	if err != nil {
		return fmt.Errorf("MFA versioning request failed")
	}
	return nil
}
