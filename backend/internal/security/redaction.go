package security

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

type redactionRule struct {
	pattern     *regexp.Regexp
	replacement string
}

var secretPatterns = []redactionRule{
	{regexp.MustCompile(`(?i)(password|passphrase|token|secret|keyring|client_key|authorization|access_key|secret_key)(\s*[=:]\s*)([^\s,;]+)`), `${1}${2}[REDACTED]`},
	{regexp.MustCompile(`(?i)(client\.[a-z0-9_.-]+\s*\{[^}]*?\bkey\s*=\s*)([^}\s]+)`), `${1}[REDACTED]`},
	{regexp.MustCompile(`(?i)(https?://)[^/@\s]+@`), `${1}[REDACTED]@`},
	{regexp.MustCompile(`\beyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\b`), `[REDACTED]`},
}

var secretFieldPattern = regexp.MustCompile(`(?i)(^|_)(password|passphrase|token|secret|client_key|keyring|authorization|private_key|access_key|secret_key|credential|certificate_key)($|_)`)

// Redact removes common credential forms before values enter logs or events.
func Redact(value string) string {
	redacted := value
	for _, rule := range secretPatterns {
		redacted = rule.pattern.ReplaceAllString(redacted, rule.replacement)
	}
	return strings.TrimSpace(redacted)
}

// RedactJSON recursively removes secrets while preserving the JSON shape.
func RedactJSON(value any) (any, error) {
	normalized, err := normalizeJSON(value)
	if err != nil {
		return nil, err
	}
	return redactJSONValue(normalized, ""), nil
}

const encryptedJSONField = "$encrypted"

// ProtectJSON encrypts sensitive JSON fields while preserving the surrounding
// request shape so services can recover the original values in memory.
func ProtectJSON(value any, databaseEncryptionKey string) (any, error) {
	normalized, err := normalizeJSON(value)
	if err != nil {
		return nil, err
	}
	return protectJSONValue(normalized, "", databaseEncryptionKey)
}

// UnprotectJSON reverses ProtectJSON and rejects malformed encrypted envelopes.
func UnprotectJSON(value any, databaseEncryptionKey string) (any, error) {
	normalized, err := normalizeJSON(value)
	if err != nil {
		return nil, err
	}
	return unprotectJSONValue(normalized, databaseEncryptionKey)
}

func normalizeJSON(value any) (any, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return nil, fmt.Errorf("encode JSON value: %w", err)
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	var normalized any
	if err := decoder.Decode(&normalized); err != nil {
		return nil, fmt.Errorf("decode JSON value: %w", err)
	}
	return normalized, nil
}

func protectJSONValue(value any, field, key string) (any, error) {
	if secretFieldPattern.MatchString(field) {
		if envelope, ok := value.(map[string]any); ok && len(envelope) == 1 {
			if _, exists := envelope[encryptedJSONField].(string); exists {
				return envelope, nil
			}
		}
		encoded, err := json.Marshal(value)
		if err != nil {
			return nil, err
		}
		ciphertext, err := Encrypt(encoded, key)
		for i := range encoded {
			encoded[i] = 0
		}
		if err != nil {
			return nil, err
		}
		return map[string]any{encryptedJSONField: ciphertext}, nil
	}
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for childField, item := range typed {
			protected, err := protectJSONValue(item, childField, key)
			if err != nil {
				return nil, err
			}
			result[childField] = protected
		}
		return result, nil
	case []any:
		result := make([]any, len(typed))
		for i, item := range typed {
			protected, err := protectJSONValue(item, field, key)
			if err != nil {
				return nil, err
			}
			result[i] = protected
		}
		return result, nil
	default:
		return value, nil
	}
}

func unprotectJSONValue(value any, key string) (any, error) {
	switch typed := value.(type) {
	case map[string]any:
		if len(typed) == 1 {
			if ciphertext, ok := typed[encryptedJSONField].(string); ok {
				plaintext, err := Decrypt(ciphertext, key)
				if err != nil {
					return nil, err
				}
				defer func() {
					for i := range plaintext {
						plaintext[i] = 0
					}
				}()
				decoder := json.NewDecoder(bytes.NewReader(plaintext))
				decoder.UseNumber()
				var original any
				if err := decoder.Decode(&original); err != nil {
					return nil, fmt.Errorf("decode protected JSON field: %w", err)
				}
				return original, nil
			}
		}
		result := make(map[string]any, len(typed))
		for field, item := range typed {
			plain, err := unprotectJSONValue(item, key)
			if err != nil {
				return nil, err
			}
			result[field] = plain
		}
		return result, nil
	case []any:
		result := make([]any, len(typed))
		for i, item := range typed {
			plain, err := unprotectJSONValue(item, key)
			if err != nil {
				return nil, err
			}
			result[i] = plain
		}
		return result, nil
	default:
		return value, nil
	}
}

func redactJSONValue(value any, field string) any {
	if secretFieldPattern.MatchString(field) {
		return "[REDACTED]"
	}
	switch typed := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(typed))
		for key, item := range typed {
			result[key] = redactJSONValue(item, key)
		}
		return result
	case []any:
		result := make([]any, len(typed))
		for i, item := range typed {
			result[i] = redactJSONValue(item, field)
		}
		return result
	case string:
		return Redact(typed)
	default:
		return value
	}
}
