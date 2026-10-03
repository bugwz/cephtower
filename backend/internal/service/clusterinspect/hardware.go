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

func (s *Service) Hardware(ctx context.Context, clusterID uint64, host, category string) (map[string]any, error) {
	if clusterID == 0 || !regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,255}$`).MatchString(host) {
		return nil, invalid("invalid hardware host or cluster id")
	}
	switch category {
	case "memory", "storage", "processors", "network", "power", "fans":
	default:
		return nil, invalid("invalid hardware category")
	}
	var report map[string]map[string]map[string]json.RawMessage
	if err := s.read(ctx, clusterID, "host.hardware", []string{"orch", "hardware", "status", "--hostname", host, "--category", category, "--format", "json"}, &report); err != nil {
		return nil, err
	}
	bad := func() (map[string]any, error) {
		return nil, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "invalid hardware status response"}
	}
	if report == nil {
		return bad()
	}
	items := make([]map[string]any, 0)
	for hostname, systems := range report {
		if hostname != host || systems == nil {
			return bad()
		}
		for system, components := range systems {
			if system == "" || components == nil {
				return bad()
			}
			for component, raw := range components {
				var details *struct {
					Status *struct {
						Health *string `json:"health"`
						State  *string `json:"state"`
					} `json:"status"`
				}
				if component == "" || json.Unmarshal(raw, &details) != nil || details == nil {
					return bad()
				}
				var health, state any
				if details.Status != nil {
					if details.Status.Health != nil {
						health = security.Redact(*details.Status.Health)
					}
					if details.Status.State != nil {
						state = security.Redact(*details.Status.State)
					}
				}
				identity, _ := json.Marshal([]string{system, component})
				items = append(items, map[string]any{"id": string(identity), "system": system, "component": component, "health": health, "state": state, "details": security.Redact(string(raw))})
			}
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i]["id"].(string) < items[j]["id"].(string) })
	return map[string]any{"host": host, "category": category, "items": items, "observed_at": time.Now().UTC()}, nil
}
