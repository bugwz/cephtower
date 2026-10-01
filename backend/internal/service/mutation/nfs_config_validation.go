package mutation

import (
	"strings"
	"unicode"
)

func requiredNFSPath(parameters map[string]any, field string) (string, error) {
	value, ok := parameters[field].(string)
	value = strings.TrimSpace(value)
	if !ok || value == "" || len(value) > 4096 {
		return "", invalid(field + " must be a non-empty NFS path of at most 4096 bytes")
	}
	if err := validateNFSConfigStrings(map[string]any{field: value}); err != nil {
		return "", err
	}
	return value, nil
}

// Ganesha's reference serializer quotes these values without escaping them.
// JSON encoding protects transport, but does not protect the generated config.
func validateNFSConfigStrings(parameters map[string]any) error {
	for _, field := range []string{"cluster", "pseudo", "path", "filesystem", "rgw_user_id"} {
		value, exists := parameters[field]
		if !exists {
			continue
		}
		text, ok := value.(string)
		if !ok {
			return invalid(field + " must be a string")
		}
		if strings.ContainsAny(text, "\"\\") || strings.IndexFunc(text, unicode.IsControl) >= 0 {
			return invalid(field + " cannot contain quotes, backslashes or control characters in NFS configuration")
		}
	}
	return nil
}
