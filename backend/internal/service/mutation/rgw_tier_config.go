package mutation

import (
	"encoding/json"
	"strings"
)

// Native tier-config scans commas and brace depth before JSONFormattable parses
// each value as JSON. Encode both layers, even for JSON-looking identity strings.
func tierConfigString(value string) string {
	raw, _ := json.Marshal(value)
	return strings.NewReplacer(",", `\u002c`, "{", `\u007b`, "}", `\u007d`).Replace(string(raw))
}
