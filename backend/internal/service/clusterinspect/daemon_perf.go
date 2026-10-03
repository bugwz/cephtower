package clusterinspect

import (
	"context"
	"encoding/json"
	"regexp"
	"sort"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/security"
)

func (s *Service) DaemonPerf(ctx context.Context, clusterID uint64, name string) (map[string]any, error) {
	if !regexp.MustCompile(`^(mon|mgr|mds|osd|rgw|rbd-mirror)\.[A-Za-z0-9_][A-Za-z0-9_.-]*$`).MatchString(name) {
		return nil, invalid("unsupported daemon identity for performance inspection")
	}
	var schema, values map[string]map[string]json.RawMessage
	if err := s.read(ctx, clusterID, "daemon.perf.schema", []string{"tell", name, "perf", "schema", "--format", "json"}, &schema); err != nil {
		return nil, err
	}
	if err := s.read(ctx, clusterID, "daemon.perf.dump", []string{"tell", name, "perf", "dump", "--format", "json"}, &values); err != nil {
		return nil, err
	}
	bad := func() (map[string]any, error) {
		return nil, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "invalid daemon performance response"}
	}
	if schema == nil || values == nil {
		return bad()
	}
	items := make([]map[string]any, 0)
	for group, counters := range schema {
		if counters == nil {
			return bad()
		}
		for counter, raw := range counters {
			var definition *struct {
				Description string `json:"description"`
				Units       string `json:"units"`
				ValueType   string `json:"value_type"`
				Type        *int   `json:"type"`
				Priority    *int   `json:"priority"`
			}
			if json.Unmarshal(raw, &definition) != nil || definition == nil {
				return bad()
			}
			var value any
			if data, exists := values[group][counter]; exists && string(data) != "null" {
				value = security.Redact(string(data)) // Keep exact integer and average-pair values as text.
			}
			items = append(items, map[string]any{"name": group + "." + counter, "description": security.Redact(definition.Description), "units": definition.Units, "value_type": definition.ValueType, "type": definition.Type, "priority": definition.Priority, "raw_value": value})
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i]["name"].(string) < items[j]["name"].(string) })
	return map[string]any{"items": items, "daemon_name": name, "observed_at": time.Now().UTC()}, nil
}
