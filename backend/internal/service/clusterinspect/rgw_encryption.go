package clusterinspect

import (
	"context"
	"net/url"
	"regexp"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
)

type RGWEncryptionField struct {
	Name     string `json:"name"`
	Option   string `json:"option"`
	Value    string `json:"value"`
	Redacted bool   `json:"redacted"`
}

type RGWEncryptionConfiguration struct {
	Entity         string               `json:"entity"`
	EncryptionType string               `json:"encryption_type"`
	Provider       string               `json:"provider"`
	Backend        string               `json:"backend"`
	Fields         []RGWEncryptionField `json:"fields"`
	ObservedAt     time.Time            `json:"observed_at"`
}

var rgwEncryptionEntity = regexp.MustCompile(`^client\.rgw\.[A-Za-z0-9][A-Za-z0-9_.-]{0,255}$`)

// Read monitor-resolved configuration, not the running daemon's effective config.
// An explicit profile can exist without being selected by its backend option.
func (s *Service) RGWEncryptionConfiguration(ctx context.Context, clusterID uint64, entity, encryptionType, provider string) (RGWEncryptionConfiguration, error) {
	var empty RGWEncryptionConfiguration
	if clusterID == 0 || !rgwEncryptionEntity.MatchString(entity) || !(encryptionType == "kms" && oneOf(provider, "vault", "kmip") || encryptionType == "s3" && provider == "vault") {
		return empty, invalid("an RGW config entity and supported encryption profile are required")
	}
	access, err := s.clusters.Access(ctx, clusterID)
	if err != nil {
		return empty, err
	}
	failed := func(code string) error {
		return &cephdomain.ActionError{Code: code, Message: "RGW encryption configuration could not be verified"}
	}
	read := func(option string) (string, error) {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: "rgw.encryption.read", Binary: executor.BinaryCeph, Args: []string{"config", "get", entity, option}, Timeout: 20 * time.Second, MaxOutput: 16384})
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		if err != nil || result.ExitCode != 0 {
			return "", failed("ceph_command_failed")
		}
		// ConfigMonitor appends exactly one newline even for empty string options.
		if len(result.Stdout) > 16384 || !utf8.Valid(result.Stdout) || !strings.HasSuffix(string(result.Stdout), "\n") {
			return "", failed("invalid_ceph_response")
		}
		value := strings.TrimSuffix(string(result.Stdout), "\n")
		if strings.ContainsFunc(value, unicode.IsControl) {
			return "", failed("invalid_ceph_response")
		}
		return value, nil
	}
	backendKey, prefix := "rgw_crypt_s3_kms_backend", "rgw_crypt_"+provider+"_"
	if encryptionType == "s3" {
		backendKey = "rgw_crypt_sse_s3_backend"
		prefix = "rgw_crypt_sse_s3_vault_"
	}
	backend, err := read(backendKey)
	if err != nil {
		return empty, err
	}
	fields := []string{"addr", "auth", "prefix", "secret_engine", "namespace", "token_file", "ssl_cacert", "ssl_clientcert", "ssl_clientkey", "verify_ssl"}
	if provider == "kmip" {
		fields = []string{"addr", "username", "password", "client_cert", "client_key", "ca_path", "kms_key_template", "s3_key_template"}
	}
	configuration := RGWEncryptionConfiguration{Entity: entity, EncryptionType: encryptionType, Provider: provider, Backend: backend, Fields: []RGWEncryptionField{}}
	for _, name := range fields {
		option := prefix + name
		value, err := read(option)
		if err != nil {
			return empty, err
		}
		field := RGWEncryptionField{Name: name, Option: option, Value: value}
		if name == "password" {
			field.Value = "[REDACTED]"
			field.Redacted = true
		}
		if name == "addr" {
			if provider == "vault" && value != "" {
				u, err := url.Parse(value)
				if err != nil || u.Hostname() == "" || !oneOf(u.Scheme, "http", "https") || u.Opaque != "" || strings.Contains(value, "\\") {
					field.Value = "[REDACTED]"
					field.Redacted = true
				} else {
					field.Redacted = u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != ""
					u.User = nil
					u.RawQuery = ""
					u.ForceQuery = false
					u.Fragment = ""
					u.RawFragment = ""
					field.Value = u.String()
				}
			} else if strings.ContainsAny(value, "@?#") {
				field.Value = "[REDACTED]"
				field.Redacted = true
			}
		}
		configuration.Fields = append(configuration.Fields, field)
	}
	// SSE-S3's key template is outside the vault prefix.
	if encryptionType == "s3" {
		value, err := read("rgw_crypt_sse_s3_key_template")
		if err != nil {
			return empty, err
		}
		configuration.Fields = append(configuration.Fields, RGWEncryptionField{Name: "key_template", Option: "rgw_crypt_sse_s3_key_template", Value: value})
	}
	check, err := read(backendKey)
	if err != nil {
		return empty, err
	}
	if check != backend {
		return empty, failed("configuration_changed")
	}
	configuration.ObservedAt = time.Now().UTC()
	return configuration, nil
}
