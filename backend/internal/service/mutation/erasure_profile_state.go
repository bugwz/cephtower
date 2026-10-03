package mutation

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

// Compare the emitted parameters, allowing Ceph to add plugin defaults.
func erasureProfileCreated(args []string, data []byte) bool {
	profile := make(map[string]string)
	decoder := json.NewDecoder(bytes.NewReader(data))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return false
	}
	for decoder.More() {
		token, err := decoder.Token()
		key, ok := token.(string)
		if err != nil || !ok || strings.TrimSpace(key) == "" {
			return false
		}
		if _, duplicate := profile[key]; duplicate {
			return false
		}
		var value *string
		if decoder.Decode(&value) != nil || value == nil {
			return false
		}
		profile[key] = *value
	}
	end, err := decoder.Token()
	if err != nil || end != json.Delim('}') || decoder.Decode(new(any)) != io.EOF || len(profile) == 0 || strings.TrimSpace(profile["plugin"]) == "" {
		return false
	}
	if len(args) < 4 || args[0] != "osd" || args[1] != "erasure-code-profile" || args[2] != "set" {
		return false
	}
	for _, arg := range args[4:] {
		key, value, ok := strings.Cut(arg, "=")
		if !ok || profile[key] != value {
			return false
		}
	}
	return true
}
