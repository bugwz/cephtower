package settings

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"cephtower/backend/internal/store"
)

type CephClient interface {
	Raw(context.Context, string, string, url.Values, any) (json.RawMessage, error)
}

type Service struct {
	ceph     CephClient
	database func() *store.Database
}

func New(ceph CephClient, database func() *store.Database) *Service {
	return &Service{ceph: ceph, database: database}
}

var groups = map[string][]string{
	"general":         {"ENABLE_BROWSABLE_API", "REST_REQUESTS_TIMEOUT", "UNSAFE_TLS_v1_2"},
	"audit":           {"AUDIT_API_ENABLED", "AUDIT_API_LOG_PAYLOAD"},
	"rgw":             {"RGW_API_ACCESS_KEY", "RGW_API_SECRET_KEY", "RGW_API_ADMIN_RESOURCE", "RGW_API_SSL_VERIFY", "RGW_HOSTNAME_PER_DAEMON"},
	"grafana":         {"GRAFANA_API_URL", "GRAFANA_FRONTEND_API_URL", "GRAFANA_API_USERNAME", "GRAFANA_API_PASSWORD", "GRAFANA_API_SSL_VERIFY", "GRAFANA_UPDATE_DASHBOARDS"},
	"prometheus":      {"PROMETHEUS_API_HOST", "PROMETHEUS_API_SSL_VERIFY", "ALERTMANAGER_API_HOST", "ALERTMANAGER_API_SSL_VERIFY", "PROM_ALERT_CREDENTIAL_CACHE_TTL"},
	"iscsi":           {"ISCSI_API_SSL_VERIFICATION"},
	"nfs":             {"GANESHA_CLUSTERS_RADOS_POOL_NAMESPACE"},
	"user-policy":     {"USER_PWD_EXPIRATION_SPAN", "USER_PWD_EXPIRATION_WARNING_1", "USER_PWD_EXPIRATION_WARNING_2"},
	"password-policy": {"PWD_POLICY_ENABLED", "PWD_POLICY_CHECK_LENGTH_ENABLED", "PWD_POLICY_CHECK_OLDPWD_ENABLED", "PWD_POLICY_CHECK_USERNAME_ENABLED", "PWD_POLICY_CHECK_EXCLUSION_LIST_ENABLED", "PWD_POLICY_CHECK_COMPLEXITY_ENABLED", "PWD_POLICY_CHECK_SEQUENTIAL_CHARS_ENABLED", "PWD_POLICY_CHECK_REPETITIVE_CHARS_ENABLED", "PWD_POLICY_MIN_LENGTH", "PWD_POLICY_MIN_COMPLEXITY", "PWD_POLICY_EXCLUSION_LIST"},
	"multi-cluster":   {"MULTICLUSTER_CONFIG", "MANAGED_BY_CLUSTERS"},
	"feedback":        {"ISSUE_TRACKER_API_KEY"},
}

func (s *Service) List(ctx context.Context, query url.Values) (any, error) {
	payload, err := s.ceph.Raw(ctx, http.MethodGet, "/api/settings", query, nil)
	if err != nil {
		return nil, err
	}
	return RedactSettings(payload), nil
}

func (s *Service) Get(ctx context.Context, name string) (any, error) {
	name = NormalizeName(name)
	payload, err := s.ceph.Raw(ctx, http.MethodGet, "/api/settings/"+url.PathEscape(name), nil, nil)
	if err != nil {
		return nil, err
	}
	return RedactSetting(name, payload), nil
}

func (s *Service) UpdateAll(ctx context.Context, query url.Values, body any, operatorID uint) (any, error) {
	updates := AuditUpdates(body)
	old := s.auditValues(ctx, updates)
	payload, err := s.ceph.Raw(ctx, http.MethodPut, "/api/settings", query, body)
	if err != nil {
		for name, value := range updates {
			s.record(ctx, operatorID, name, old[name], JSON(RedactValue(name, value)), "failed", err.Error())
		}
		return nil, err
	}
	for name := range updates {
		s.record(ctx, operatorID, name, old[name], s.auditValue(ctx, name), "success", "")
	}
	return RedactSettings(payload), nil
}

func (s *Service) Update(ctx context.Context, name string, value any, operatorID uint) (any, error) {
	name = NormalizeName(name)
	old := s.auditValue(ctx, name)
	body := map[string]any{"value": value}
	payload, err := s.ceph.Raw(ctx, http.MethodPut, "/api/settings/"+url.PathEscape(name), nil, body)
	if err != nil {
		s.record(ctx, operatorID, name, old, JSON(RedactValue(name, value)), "failed", err.Error())
		return nil, err
	}
	s.record(ctx, operatorID, name, old, s.auditValue(ctx, name), "success", "")
	return RedactSetting(name, payload), nil
}

func (s *Service) Reset(ctx context.Context, name string) (any, error) {
	name = NormalizeName(name)
	payload, err := s.ceph.Raw(ctx, http.MethodDelete, "/api/settings/"+url.PathEscape(name), nil, nil)
	if err != nil {
		return nil, err
	}
	return RedactSetting(name, payload), nil
}

func (s *Service) ListGroups(ctx context.Context) ([]Group, error) {
	settings, err := s.listNormalized(ctx)
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(groups))
	for name := range groups {
		names = append(names, name)
	}
	sort.Strings(names)
	result := make([]Group, 0, len(names))
	for _, name := range names {
		result = append(result, Group{Name: name, Settings: forGroup(settings, groups[name])})
	}
	return result, nil
}

func (s *Service) GetGroup(ctx context.Context, group string) (Group, error) {
	group = NormalizeGroup(group)
	names, ok := groups[group]
	if !ok {
		return Group{}, ErrGroupNotFound
	}
	settings, err := s.listNormalized(ctx)
	if err != nil {
		return Group{}, err
	}
	return Group{Name: group, Settings: forGroup(settings, names)}, nil
}

func (s *Service) UpdateGroup(ctx context.Context, group string, updates map[string]any, operatorID uint) (Group, error) {
	group = NormalizeGroup(group)
	names, ok := groups[group]
	if !ok {
		return Group{}, ErrGroupNotFound
	}
	allowed := map[string]bool{}
	for _, name := range names {
		allowed[NormalizeName(name)] = true
	}
	for name, value := range updates {
		normalized := NormalizeName(name)
		if !allowed[normalized] {
			return Group{}, fmt.Errorf("%w: %s", ErrInvalidGroup, name)
		}
		if _, err := s.Update(ctx, normalized, value, operatorID); err != nil {
			return Group{}, err
		}
	}
	return s.GetGroup(ctx, group)
}

func (s *Service) listNormalized(ctx context.Context) ([]Setting, error) {
	payload, err := s.ceph.Raw(ctx, http.MethodGet, "/api/settings", nil, nil)
	if err != nil {
		return nil, err
	}
	result, ok := RedactSettings(payload).([]Setting)
	if !ok {
		return nil, ErrInvalidReply
	}
	return result, nil
}

func (s *Service) auditValues(ctx context.Context, updates map[string]any) map[string]string {
	result := map[string]string{}
	for name := range updates {
		result[name] = s.auditValue(ctx, name)
	}
	return result
}

func (s *Service) auditValue(ctx context.Context, name string) string {
	payload, err := s.ceph.Raw(ctx, http.MethodGet, "/api/settings/"+url.PathEscape(name), nil, nil)
	if err != nil {
		return ""
	}
	if setting, ok := RedactSetting(name, payload).(Setting); ok {
		return JSON(setting.Value)
	}
	return JSON(RedactAny(payload))
}

func (s *Service) record(ctx context.Context, operatorID uint, name, oldValue, newValue, status, message string) {
	if s.database == nil || s.database() == nil {
		return
	}
	cluster, err := s.database().FirstCluster(ctx)
	if err != nil {
		return
	}
	_ = s.database().RecordSettingChange(ctx, &store.CephClusterSettingChange{ClusterID: cluster.ID, SettingName: NormalizeName(name), OldValueRedacted: oldValue, NewValueRedacted: newValue, OperatorUserID: operatorID, Source: "api", Status: status, Error: message})
}

func RedactSettings(payload json.RawMessage) any {
	var list []map[string]any
	if json.Unmarshal(payload, &list) == nil {
		result := make([]Setting, 0, len(list))
		for _, item := range list {
			result = append(result, normalize(item))
		}
		return result
	}
	return RedactAny(payload)
}

func RedactSetting(fallback string, payload json.RawMessage) any {
	var item map[string]any
	if json.Unmarshal(payload, &item) == nil {
		if item["name"] == nil {
			item["name"] = fallback
		}
		return normalize(item)
	}
	return RedactAny(payload)
}

func normalize(item map[string]any) Setting {
	name := NormalizeName(text(item["name"]))
	value, set := item["value"]
	return Setting{Name: name, Type: text(item["type"]), Default: boolean(item["default"]), Sensitive: Sensitive(name), ValueSet: set && value != nil && text(value) != "", Value: RedactValue(name, value)}
}

func RedactAny(payload json.RawMessage) any {
	var value any
	if json.Unmarshal(payload, &value) != nil {
		return nil
	}
	return redact("", value)
}
func redact(name string, value any) any {
	switch value := value.(type) {
	case map[string]any:
		result := make(map[string]any, len(value))
		for key, child := range value {
			if Sensitive(key) {
				result[key] = scalar(child)
			} else {
				result[key] = redact(key, child)
			}
		}
		return result
	case []any:
		result := make([]any, 0, len(value))
		for _, child := range value {
			result = append(result, redact(name, child))
		}
		return result
	default:
		return RedactValue(name, value)
	}
}
func RedactValue(name string, value any) any {
	if Sensitive(name) {
		return scalar(value)
	}
	return value
}
func scalar(value any) any {
	if value == nil || text(value) == "" {
		return nil
	}
	return "********"
}
func Sensitive(name string) bool {
	upper := strings.ToUpper(name)
	for _, marker := range []string{"PASSWORD", "SECRET", "ACCESS_KEY", "API_KEY", "TOKEN", "KEYRING", "CREDENTIAL"} {
		if strings.Contains(upper, marker) {
			return true
		}
	}
	return false
}
func NormalizeName(name string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(name), "-", "_"))
}
func NormalizeGroup(name string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(name), "_", "-"))
}
func JSON(value any) string {
	data, err := json.Marshal(value)
	if err != nil {
		return ""
	}
	return string(data)
}
func text(value any) string {
	if value == nil {
		return ""
	}
	if s, ok := value.(string); ok {
		return s
	}
	if n, ok := value.(json.Number); ok {
		return n.String()
	}
	return strings.Trim(strings.TrimSpace(JSON(value)), `"`)
}
func boolean(value any) bool {
	if b, ok := value.(bool); ok {
		return b
	}
	return strings.EqualFold(text(value), "true")
}
func forGroup(settings []Setting, names []string) []Setting {
	byName := map[string]Setting{}
	for _, item := range settings {
		byName[NormalizeName(item.Name)] = item
	}
	result := make([]Setting, 0, len(names))
	for _, name := range names {
		if item, ok := byName[NormalizeName(name)]; ok {
			result = append(result, item)
		}
	}
	return result
}
func AuditUpdates(body any) map[string]any {
	result := map[string]any{}
	raw, ok := body.(json.RawMessage)
	if !ok {
		return result
	}
	var object map[string]any
	if json.Unmarshal(raw, &object) != nil {
		return result
	}
	if nested, ok := object["settings"].(map[string]any); ok {
		object = nested
	}
	if nested, ok := object["values"].(map[string]any); ok {
		object = nested
	}
	for key, value := range object {
		name := NormalizeName(key)
		if item, ok := value.(map[string]any); ok {
			if nested, exists := item["value"]; exists {
				result[name] = nested
				continue
			}
		}
		result[name] = value
	}
	return result
}
