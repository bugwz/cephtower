package s3

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
)

func ValidateBucketConfiguration(kind string, body []byte) error {
	if len(body) == 0 || len(body) > 4<<20 {
		return fmt.Errorf("bucket configuration must be between 1 byte and 4 MiB")
	}
	if kind == "policy" {
		var document map[string]json.RawMessage
		if json.Unmarshal(body, &document) != nil || document == nil {
			return fmt.Errorf("bucket policy must be a JSON object")
		}
		return nil
	}
	root := map[string]string{"cors": "CORSConfiguration", "lifecycle": "LifecycleConfiguration", "encryption": "ServerSideEncryptionConfiguration", "versioning": "VersioningConfiguration"}[kind]
	if root == "" {
		return fmt.Errorf("unsupported S3 bucket configuration %q", kind)
	}
	decoder := xml.NewDecoder(bytes.NewReader(body))
	depth, roots := 0, 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("bucket configuration must be well-formed XML")
		}
		switch value := token.(type) {
		case xml.StartElement:
			if depth == 0 {
				roots++
				if roots != 1 || value.Name.Local != root {
					return fmt.Errorf("bucket configuration XML root must be %s", root)
				}
			}
			depth++
		case xml.EndElement:
			depth--
		case xml.CharData:
			if depth == 0 && len(bytes.TrimSpace(value)) != 0 {
				return fmt.Errorf("text outside configuration XML root is not allowed")
			}
		case xml.Directive:
			return fmt.Errorf("XML directives are not allowed")
		case xml.ProcInst:
			if value.Target != "xml" || roots != 0 {
				return fmt.Errorf("XML processing instructions are not allowed")
			}
		}
	}
	if roots != 1 || depth != 0 {
		return fmt.Errorf("bucket configuration XML root is required")
	}
	return nil
}
