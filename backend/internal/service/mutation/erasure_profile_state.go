package mutation

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
)

// Compare the emitted parameters, allowing Ceph to add plugin defaults.
func erasureProfileCreated(args []string, data []byte) bool {
	var profile map[string]string
	decoder := json.NewDecoder(bytes.NewReader(data))
	if decoder.Decode(&profile) != nil || decoder.Decode(new(any)) != io.EOF || len(profile) == 0 || profile["plugin"] == "" {
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
