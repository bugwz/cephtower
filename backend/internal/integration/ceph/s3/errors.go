package s3

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// ResponseError retains the native S3 status and a bounded response body.
type ResponseError struct {
	StatusCode int
	Code       string
	message    string
}

func (e *ResponseError) Error() string {
	return fmt.Sprintf("S3 returned HTTP %d: %s", e.StatusCode, e.message)
}

func responseError(status int, body io.Reader) error {
	const limit = 32 << 10
	data, err := io.ReadAll(io.LimitReader(body, limit+1))
	code := ""
	if err == nil && len(data) <= limit {
		code = errorCode(data)
	}
	if len(data) > limit {
		data = data[:limit]
	}
	return &ResponseError{StatusCode: status, Code: code, message: strings.TrimSpace(string(data))}
}

// Parse the complete envelope: malformed, duplicate or truncated codes are unknown.
func errorCode(data []byte) string {
	d := xml.NewDecoder(bytes.NewReader(data))
	depth, roots, codes := 0, 0, 0
	code := ""
	for {
		token, err := d.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return ""
		}
		switch v := token.(type) {
		case xml.StartElement:
			if depth == 0 {
				roots++
				if roots != 1 || v.Name.Local != "Error" {
					return ""
				}
			}
			if depth == 1 && v.Name.Local == "Code" {
				codes++
				var value struct {
					Text     string     `xml:",chardata"`
					Children []xml.Name `xml:",any"`
				}
				if d.DecodeElement(&value, &v) != nil || len(value.Children) != 0 {
					return ""
				}
				code = value.Text
				continue
			}
			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 0 && strings.TrimSpace(string(v)) != "" {
				return ""
			}
		case xml.Directive:
			return ""
		}
	}
	if roots != 1 || depth != 0 || codes != 1 {
		return ""
	}
	return code
}

func IsConfigurationMissing(kind string, err error) bool {
	codes := map[string]string{"policy": "NoSuchBucketPolicy", "cors": "NoSuchCORSConfiguration", "lifecycle": "NoSuchLifecycleConfiguration", "encryption": "ServerSideEncryptionConfigurationNotFoundError", "tagging": "NoSuchTagSet", "object-lock": "ObjectLockConfigurationNotFoundError"}
	var response *ResponseError
	return codes[kind] != "" && errors.As(err, &response) && response.StatusCode == http.StatusNotFound && response.Code == codes[kind]
}
