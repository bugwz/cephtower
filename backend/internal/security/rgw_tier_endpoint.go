package security

import (
	"net/url"
	"strings"
	"unicode"
)

// Redact only the display projection. Never use it to build a native mutation.
func redactTierEndpoint(value any) (string, bool) {
	raw, ok := value.(string)
	if !ok {
		return "[REDACTED]", true
	}
	if raw == "" {
		return "", false
	}
	if strings.Contains(raw, `\`) || strings.IndexFunc(raw, func(r rune) bool { return unicode.IsControl(r) || unicode.IsSpace(r) }) >= 0 {
		return "[REDACTED]", true
	}
	parsed, err := url.Parse(raw)
	if err != nil || parsed.Hostname() == "" || parsed.Opaque != "" || parsed.OmitHost || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "[REDACTED]", true
	}
	hidden := parsed.User != nil || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Fragment != "" || parsed.RawFragment != ""
	parsed.User = nil
	parsed.RawQuery = ""
	parsed.ForceQuery = false
	parsed.Fragment = ""
	parsed.RawFragment = ""
	return Redact(parsed.String()), hidden
}
