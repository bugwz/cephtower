package mutation

import (
	"encoding/base64"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type rgwEncryptionChange struct {
	field, option, value string
	command              command
}

type rgwEncryptionPlan struct {
	entity, encryptionType, provider, backendKey, expectedBackend string
	changes                                                       []rgwEncryptionChange
	options                                                       []string
}

var rgwEncryptionEntityPattern = regexp.MustCompile(`^client\.rgw\.[A-Za-z0-9][A-Za-z0-9_.-]{0,255}$`)

// Provider fields are explicit patches. Backend selection is never implied by
// choosing which provider to configure, and omitted fields remain untouched.
func planRGWEncryption(p map[string]any) (rgwEncryptionPlan, error) {
	var empty rgwEncryptionPlan
	fail := func() (rgwEncryptionPlan, error) {
		return empty, invalid("invalid RGW encryption configuration patch or missing confirmation")
	}
	for key := range p {
		switch key {
		case "entity", "encryption_type", "provider", "expected_backend", "values", "confirm_disruption", "confirm_credentials_saved":
		default:
			return fail()
		}
	}
	entity, _ := p["entity"].(string)
	encryptionType, _ := p["encryption_type"].(string)
	provider, _ := p["provider"].(string)
	backend, _ := p["expected_backend"].(string)
	if !rgwEncryptionEntityPattern.MatchString(entity) || p["confirm_disruption"] != true || backend == "" || len(backend) > 128 || !utf8.ValidString(backend) || strings.ContainsFunc(backend, func(r rune) bool { return unicode.IsSpace(r) || unicode.IsControl(r) }) || !(encryptionType == "kms" && (provider == "vault" || provider == "kmip") || encryptionType == "s3" && provider == "vault") {
		return fail()
	}
	if confirmation, present := p["confirm_credentials_saved"]; present {
		if _, ok := confirmation.(bool); !ok {
			return fail()
		}
	}
	values, ok := p["values"].(map[string]any)
	if !ok || len(values) == 0 {
		return fail()
	}
	allowed := map[string]bool{}
	fields := []string{"addr", "auth", "prefix", "secret_engine", "namespace", "token_file", "ssl_cacert", "ssl_clientcert", "ssl_clientkey", "verify_ssl"}
	if provider == "kmip" {
		fields = []string{"addr", "username", "password", "client_cert", "client_key", "ca_path", "kms_key_template", "s3_key_template"}
	}
	if encryptionType == "s3" {
		fields = append(fields, "key_template")
	}
	for _, field := range fields {
		allowed[field] = true
	}
	plan := rgwEncryptionPlan{entity: entity, encryptionType: encryptionType, provider: provider, backendKey: "rgw_crypt_s3_kms_backend", expectedBackend: backend}
	prefix := "rgw_crypt_" + provider + "_"
	if encryptionType == "s3" {
		prefix = "rgw_crypt_sse_s3_vault_"
		plan.backendKey = "rgw_crypt_sse_s3_backend"
	}
	plan.options = append(plan.options, plan.backendKey)
	for _, field := range fields {
		option := prefix + field
		if field == "key_template" {
			option = "rgw_crypt_sse_s3_key_template"
		}
		plan.options = append(plan.options, option)
	}
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		if !allowed[name] {
			return fail()
		}
		value, ok := values[name].(string)
		if name == "verify_ssl" {
			b, boolean := values[name].(bool)
			if !boolean {
				return fail()
			}
			value = strconv.FormatBool(b)
		} else if !ok {
			return fail()
		}
		if len(value) > 4096 || !utf8.ValidString(value) || strings.ContainsFunc(value, unicode.IsControl) || strings.Contains(value, "[REDACTED]") {
			return fail()
		}
		if name == "password" && p["confirm_credentials_saved"] != true {
			return fail()
		}
		if name == "auth" && value != "token" && value != "agent" {
			return fail()
		}
		if name == "secret_engine" && value != "transit" && (encryptionType != "kms" || value != "kv-v2") {
			return fail()
		}
		if name == "addr" && value != "" {
			if provider == "vault" {
				if !cloudConnectionEndpoint(value) {
					return fail()
				}
			} else {
				// This reference client splits at the first colon, so IPv6 is not
				// accepted here as if it were a supported host/port representation.
				if strings.ContainsAny(value, "/@?#\\[]") || strings.ContainsFunc(value, unicode.IsSpace) || strings.Count(value, ":") > 1 {
					return fail()
				}
				host, port, hasPort := strings.Cut(value, ":")
				if host == "" {
					return fail()
				}
				if hasPort {
					number, err := strconv.Atoi(port)
					if err != nil || number < 1 || number > 65535 || strconv.Itoa(number) != port {
						return fail()
					}
				}
			}
		}
		option := prefix + name
		if name == "key_template" {
			option = "rgw_crypt_sse_s3_key_template"
		}
		resource := "configuration/value/" + base64.RawURLEncoding.EncodeToString([]byte(entity+"\x00"+option))
		cmd, err := configurationCommand(Request{Action: "config_value.set", ResourceKey: resource}, map[string]any{"value": value})
		if err != nil {
			return fail()
		}
		// Do not rely on a field-name heuristic for encryption configuration values.
		cmd.sensitive = map[int]struct{}{len(cmd.args) - 1: {}}
		plan.changes = append(plan.changes, rgwEncryptionChange{field: name, option: option, value: value, command: cmd})
	}
	return plan, nil
}
