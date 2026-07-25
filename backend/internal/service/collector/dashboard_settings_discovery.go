package collector

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"cephtower/backend/internal/store"
)

var settingSnapshotGroups = map[string][]string{
	"general": {
		"ENABLE_BROWSABLE_API",
		"REST_REQUESTS_TIMEOUT",
		"UNSAFE_TLS_v1_2",
	},
	"audit": {
		"AUDIT_API_ENABLED",
		"AUDIT_API_LOG_PAYLOAD",
	},
	"rgw": {
		"RGW_API_ACCESS_KEY",
		"RGW_API_SECRET_KEY",
		"RGW_API_ADMIN_RESOURCE",
		"RGW_API_SSL_VERIFY",
		"RGW_HOSTNAME_PER_DAEMON",
	},
	"grafana": {
		"GRAFANA_API_URL",
		"GRAFANA_FRONTEND_API_URL",
		"GRAFANA_API_USERNAME",
		"GRAFANA_API_PASSWORD",
		"GRAFANA_API_SSL_VERIFY",
		"GRAFANA_UPDATE_DASHBOARDS",
	},
	"prometheus": {
		"PROMETHEUS_API_HOST",
		"PROMETHEUS_API_SSL_VERIFY",
		"ALERTMANAGER_API_HOST",
		"ALERTMANAGER_API_SSL_VERIFY",
		"PROM_ALERT_CREDENTIAL_CACHE_TTL",
	},
	"iscsi": {
		"ISCSI_API_SSL_VERIFICATION",
	},
	"nfs": {
		"GANESHA_CLUSTERS_RADOS_POOL_NAMESPACE",
	},
	"user-policy": {
		"USER_PWD_EXPIRATION_SPAN",
		"USER_PWD_EXPIRATION_WARNING_1",
		"USER_PWD_EXPIRATION_WARNING_2",
	},
	"password-policy": {
		"PWD_POLICY_ENABLED",
		"PWD_POLICY_CHECK_LENGTH_ENABLED",
		"PWD_POLICY_CHECK_OLDPWD_ENABLED",
		"PWD_POLICY_CHECK_USERNAME_ENABLED",
		"PWD_POLICY_CHECK_EXCLUSION_LIST_ENABLED",
		"PWD_POLICY_CHECK_COMPLEXITY_ENABLED",
		"PWD_POLICY_CHECK_SEQUENTIAL_CHARS_ENABLED",
		"PWD_POLICY_CHECK_REPETITIVE_CHARS_ENABLED",
		"PWD_POLICY_MIN_LENGTH",
		"PWD_POLICY_MIN_COMPLEXITY",
		"PWD_POLICY_EXCLUSION_LIST",
	},
	"multi-cluster": {
		"MULTICLUSTER_CONFIG",
		"MANAGED_BY_CLUSTERS",
	},
	"feedback": {
		"ISSUE_TRACKER_API_KEY",
	},
}

func saveDiscoveredSettings(ctx context.Context, db *store.Database, clusterID uint, payload json.RawMessage) int {
	var items []map[string]any
	if err := json.Unmarshal(payload, &items); err != nil {
		return 0
	}

	groupByName := settingGroupLookup()
	now := time.Now()
	records := make([]store.CephClusterSettingSnapshot, 0, len(items))
	for _, item := range items {
		name := normalizeSettingName(stringField(item, "name"))
		if name == "" {
			continue
		}
		value := item["value"]
		sensitive := isSensitiveSettingName(name)
		valueSet := value != nil && strings.TrimSpace(valueAsString(value)) != ""
		records = append(records, store.CephClusterSettingSnapshot{
			ClusterID:     clusterID,
			Name:          name,
			Group:         groupByName[name],
			Type:          firstStringField(item, "type"),
			Default:       boolField(item, "default"),
			Sensitive:     sensitive,
			ValueSet:      valueSet,
			ValueRedacted: mustJSON(redactedSettingValue(name, value)),
			Payload:       mustJSON(redactSettingPayload(item)),
			DiscoveredAt:  now,
		})
	}
	replaceDiscoveredRecords(ctx, db, clusterID, &store.CephClusterSettingSnapshot{}, records)
	return len(records)
}

func saveDiscoveredFeatureToggles(ctx context.Context, db *store.Database, clusterID uint, payload json.RawMessage) int {
	var values map[string]any
	if err := json.Unmarshal(payload, &values); err != nil {
		return 0
	}

	now := time.Now()
	records := make([]store.CephClusterFeatureToggle, 0, len(values))
	for name, value := range values {
		records = append(records, store.CephClusterFeatureToggle{
			ClusterID:    clusterID,
			Name:         strings.ToLower(strings.TrimSpace(name)),
			Enabled:      truthy(value),
			Payload:      mustJSON(map[string]any{name: value}),
			DiscoveredAt: now,
		})
	}
	replaceDiscoveredRecords(ctx, db, clusterID, &store.CephClusterFeatureToggle{}, records)
	return len(records)
}

func (service Service) fetchIntegrationStatus(ctx context.Context, db *store.Database, cluster *store.CephCluster) (int, error) {
	checks := []struct {
		name string
		path string
	}{
		{"dashboard", "/api/summary"},
		{"grafana", "/api/grafana/url"},
		{"prometheus", "/api/prometheus"},
		{"alertmanager", "/api/prometheus/silences"},
		{"rgw", "/api/rgw/daemon"},
		{"telemetry", "/api/telemetry/report"},
	}

	now := time.Now()
	records := make([]store.CephClusterIntegrationStatus, 0, len(checks))
	for _, check := range checks {
		payload, err := dashboardRaw(ctx, service.workDir, cluster, http.MethodGet, check.path, nil, nil)
		status := store.CephClusterIntegrationStatus{
			ClusterID:   cluster.ID,
			Integration: check.name,
			Configured:  err == nil,
			Healthy:     err == nil,
			Message:     "ok",
			Payload:     mustJSON(json.RawMessage(payload)),
			CheckedAt:   now,
		}
		if err != nil {
			status.Configured = false
			status.Healthy = false
			status.Message = err.Error()
			status.Payload = "{}"
		}
		records = append(records, status)
	}
	replaceDiscoveredRecords(ctx, db, cluster.ID, &store.CephClusterIntegrationStatus{}, records)
	return len(records), nil
}

func settingGroupLookup() map[string]string {
	result := map[string]string{}
	for group, names := range settingSnapshotGroups {
		for _, name := range names {
			result[normalizeSettingName(name)] = group
		}
	}
	return result
}

func normalizeSettingName(name string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(name), "-", "_"))
}

func isSensitiveSettingName(name string) bool {
	upper := strings.ToUpper(name)
	for _, marker := range []string{"PASSWORD", "SECRET", "ACCESS_KEY", "API_KEY", "TOKEN", "KEYRING", "CREDENTIAL"} {
		if strings.Contains(upper, marker) {
			return true
		}
	}
	return false
}

func redactSettingPayload(item map[string]any) map[string]any {
	result := make(map[string]any, len(item)+2)
	name := normalizeSettingName(stringField(item, "name"))
	for key, value := range item {
		if strings.EqualFold(key, "value") {
			result[key] = redactedSettingValue(name, value)
			continue
		}
		result[key] = value
	}
	result["sensitive"] = isSensitiveSettingName(name)
	result["value_set"] = item["value"] != nil && strings.TrimSpace(valueAsString(item["value"])) != ""
	return result
}

func redactedSettingValue(name string, value any) any {
	if !isSensitiveSettingName(name) {
		return value
	}
	if value == nil || strings.TrimSpace(valueAsString(value)) == "" {
		return nil
	}
	return "********"
}

func valueAsString(value any) string {
	switch typed := value.(type) {
	case string:
		return typed
	case json.Number:
		return typed.String()
	default:
		if value == nil {
			return ""
		}
		return strings.Trim(strings.TrimSpace(mustJSON(value)), `"`)
	}
}

func truthy(value any) bool {
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		return strings.EqualFold(strings.TrimSpace(typed), "true") || typed == "1"
	case float64:
		return typed != 0
	case int:
		return typed != 0
	default:
		return false
	}
}
