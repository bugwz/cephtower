package endpoint

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"cephtower/backend/internal/security"
	"cephtower/backend/internal/store"
)

var allowedKinds = map[string]struct{}{
	"ca":           {},
	"alertmanager": {}, "grafana": {}, "iscsi": {}, "nvmeof": {},
	"prometheus": {}, "rgw": {}, "rgw_admin": {}, "s3": {},
}

type CredentialInput struct {
	Kind  string
	Value map[string]any
}

type EndpointInput struct {
	Kind, Name, URL, TLSMode string
	CACredentialID           *uint64
	TimeoutSeconds           int
	Enabled                  *bool
}

type Service struct {
	database      func() *store.Database
	encryptionKey string
}

func New(database func() *store.Database, encryptionKey string) *Service {
	return &Service{database: database, encryptionKey: encryptionKey}
}

func (s *Service) ListCredentials(ctx context.Context, clusterID uint64) ([]store.CephClusterCredential, error) {
	return s.database().ListCredentials(ctx, clusterID)
}
func (s *Service) PutCredential(ctx context.Context, clusterID uint64, input CredentialInput) (store.CephClusterCredential, error) {
	kind, err := validateKind(input.Kind)
	if err != nil {
		return store.CephClusterCredential{}, err
	}
	if len(input.Value) == 0 {
		return store.CephClusterCredential{}, fmt.Errorf("credential must be a non-empty object")
	}
	if err := validateCredential(kind, input.Value); err != nil {
		return store.CephClusterCredential{}, err
	}
	plain, err := json.Marshal(input.Value)
	if err != nil {
		return store.CephClusterCredential{}, fmt.Errorf("encode credential: %w", err)
	}
	encrypted, err := security.Encrypt(plain, s.encryptionKey)
	if err != nil {
		return store.CephClusterCredential{}, err
	}
	fingerprint := credentialFingerprint(s.encryptionKey, plain)
	for i := range plain {
		plain[i] = 0
	}
	now := time.Now().UTC()
	row := store.CephClusterCredential{ClusterID: clusterID, Kind: kind, Credential: encrypted, Fingerprint: fingerprint, CreatedAt: now, UpdatedAt: now}
	if err := s.database().UpsertCredential(ctx, &row); err != nil {
		return store.CephClusterCredential{}, err
	}
	return s.database().FindCredential(ctx, clusterID, kind)
}

func validateCredential(kind string, value map[string]any) error {
	allowed := map[string]map[string]bool{
		"ca":           {"ca_certificate": true},
		"prometheus":   {"token": true},
		"alertmanager": {"token": true},
		"grafana":      {"token": true},
		"iscsi":        {"username": true, "password": true},
		"nvmeof":       {"token": true, "client_certificate": true, "client_key": true},
		"s3":           {"access_key": true, "secret_key": true, "session_token": true, "region": true},
		"rgw":          {"access_key": true, "secret_key": true, "session_token": true, "region": true},
		"rgw_admin":    {"username": true, "password": true},
	}[kind]
	for name, item := range value {
		if !allowed[name] {
			return fmt.Errorf("credential kind %s contains unknown field %q", kind, name)
		}
		if _, ok := item.(string); !ok {
			return fmt.Errorf("credential field %s must be a string", name)
		}
	}
	required := map[string][]string{
		"ca": {"ca_certificate"}, "iscsi": {"username", "password"},
		"s3": {"access_key", "secret_key"}, "rgw": {"access_key", "secret_key"},
	}
	for _, name := range required[kind] {
		item, ok := value[name].(string)
		if !ok || strings.TrimSpace(item) == "" {
			return fmt.Errorf("credential field %s is required", name)
		}
	}
	if (value["client_certificate"] == nil) != (value["client_key"] == nil) {
		return fmt.Errorf("client_certificate and client_key must be configured together")
	}
	return nil
}
func (s *Service) DeleteCredential(ctx context.Context, clusterID uint64, kind string) error {
	kind, err := validateKind(kind)
	if err != nil {
		return err
	}
	return s.database().DeleteCredential(ctx, clusterID, kind)
}
func (s *Service) DecryptCredential(ctx context.Context, clusterID uint64, kind string, target any) error {
	row, err := s.database().FindCredential(ctx, clusterID, kind)
	if err != nil {
		return err
	}
	return s.decryptCredential(row, target)
}
func (s *Service) DecryptCredentialByID(ctx context.Context, clusterID, credentialID uint64, target any) error {
	rows, err := s.database().ListCredentials(ctx, clusterID)
	if err != nil {
		return err
	}
	for _, row := range rows {
		if row.ID == credentialID {
			return s.decryptCredential(row, target)
		}
	}
	return store.ErrRecordNotFound
}
func (s *Service) decryptCredential(row store.CephClusterCredential, target any) error {
	plain, err := security.Decrypt(row.Credential, s.encryptionKey)
	if err != nil {
		return err
	}
	defer func() {
		for i := range plain {
			plain[i] = 0
		}
	}()
	decoder := json.NewDecoder(strings.NewReader(string(plain)))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}
func (s *Service) ListEndpoints(ctx context.Context, clusterID uint64) ([]store.CephClusterEndpoint, error) {
	return s.database().ListEndpoints(ctx, clusterID)
}
func (s *Service) Endpoint(ctx context.Context, clusterID uint64, kind string) (store.CephClusterEndpoint, error) {
	rows, err := s.database().ListEndpoints(ctx, clusterID)
	if err != nil {
		return store.CephClusterEndpoint{}, err
	}
	for _, row := range rows {
		if row.Kind == kind && row.Enabled && row.Name == "default" {
			return row, nil
		}
	}
	for _, row := range rows {
		if row.Kind == kind && row.Enabled {
			return row, nil
		}
	}
	return store.CephClusterEndpoint{}, store.ErrRecordNotFound
}
func (s *Service) CreateEndpoint(ctx context.Context, clusterID uint64, input EndpointInput) (store.CephClusterEndpoint, error) {
	row, err := s.endpointRow(ctx, clusterID, input)
	if err != nil {
		return store.CephClusterEndpoint{}, err
	}
	if err := s.database().CreateEndpoint(ctx, &row); err != nil {
		return store.CephClusterEndpoint{}, err
	}
	return row, nil
}
func (s *Service) UpdateEndpoint(ctx context.Context, clusterID, endpointID uint64, input EndpointInput) (store.CephClusterEndpoint, error) {
	existing, err := s.database().FindEndpoint(ctx, clusterID, endpointID)
	if err != nil {
		return store.CephClusterEndpoint{}, err
	}
	row, err := s.endpointRow(ctx, clusterID, input)
	if err != nil {
		return store.CephClusterEndpoint{}, err
	}
	row.ID, row.CreatedAt = existing.ID, existing.CreatedAt
	if err := s.database().SaveEndpoint(ctx, &row); err != nil {
		return store.CephClusterEndpoint{}, err
	}
	return row, nil
}
func (s *Service) DeleteEndpoint(ctx context.Context, clusterID, endpointID uint64) error {
	row, err := s.database().FindEndpoint(ctx, clusterID, endpointID)
	if err != nil {
		return err
	}
	return s.database().DeleteEndpoint(ctx, &row)
}
func (s *Service) endpointRow(ctx context.Context, clusterID uint64, input EndpointInput) (store.CephClusterEndpoint, error) {
	kind, err := validateKind(input.Kind)
	if err != nil {
		return store.CephClusterEndpoint{}, err
	}
	name := strings.TrimSpace(input.Name)
	if name == "" {
		name = "default"
	}
	parsed, err := url.ParseRequestURI(strings.TrimSpace(input.URL))
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.User != nil {
		return store.CephClusterEndpoint{}, fmt.Errorf("url must be an http(s) URL without embedded credentials")
	}
	tlsMode := strings.TrimSpace(input.TLSMode)
	if tlsMode == "" {
		tlsMode = "verify_system"
	}
	if tlsMode != "verify_system" && tlsMode != "verify_custom_ca" {
		return store.CephClusterEndpoint{}, fmt.Errorf("tls_mode must be verify_system or verify_custom_ca")
	}
	if tlsMode == "verify_custom_ca" && input.CACredentialID == nil {
		return store.CephClusterEndpoint{}, fmt.Errorf("ca_credential_id is required for verify_custom_ca")
	}
	if input.CACredentialID != nil {
		credentials, err := s.database().ListCredentials(ctx, clusterID)
		if err != nil {
			return store.CephClusterEndpoint{}, err
		}
		found := false
		for _, credential := range credentials {
			if credential.ID == *input.CACredentialID {
				found = true
				break
			}
		}
		if !found {
			return store.CephClusterEndpoint{}, fmt.Errorf("CA credential does not belong to cluster")
		}
	}
	timeout := input.TimeoutSeconds
	if timeout == 0 {
		timeout = 10
	}
	if timeout < 1 || timeout > 120 {
		return store.CephClusterEndpoint{}, fmt.Errorf("timeout_seconds must be between 1 and 120")
	}
	config, _ := json.Marshal(map[string]int{"timeout_seconds": timeout})
	configJSON := string(config)
	enabled := true
	if input.Enabled != nil {
		enabled = *input.Enabled
	}
	now := time.Now().UTC()
	return store.CephClusterEndpoint{ClusterID: clusterID, Kind: kind, Name: name, URL: parsed.String(), TLSMode: tlsMode, CACredentialID: input.CACredentialID, ConfigJSON: &configJSON, Enabled: enabled, CreatedAt: now, UpdatedAt: now}, nil
}
func validateKind(value string) (string, error) {
	value = strings.ToLower(strings.TrimSpace(value))
	if _, ok := allowedKinds[value]; !ok {
		return "", fmt.Errorf("unsupported endpoint or credential kind %q", value)
	}
	return value, nil
}
func credentialFingerprint(key string, value []byte) string {
	mac := hmac.New(sha256.New, []byte(key))
	_, _ = mac.Write(value)
	return hex.EncodeToString(mac.Sum(nil))
}
