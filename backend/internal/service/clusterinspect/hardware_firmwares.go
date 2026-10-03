package clusterinspect

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/security"
)

func (s *Service) hardwareFirmwares(ctx context.Context, clusterID uint64, host string) (map[string]any, error) {
	var report map[string]map[string]json.RawMessage
	if err := s.read(ctx, clusterID, "host.hardware", []string{"orch", "hardware", "status", "--hostname", host, "--category", "firmwares", "--format", "json"}, &report); err != nil {
		return nil, err
	}
	bad := func() (map[string]any, error) {
		return nil, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "invalid firmware inventory response"}
	}
	if report == nil {
		return bad()
	}
	text := func(value *string) any {
		if value == nil {
			return nil
		}
		return security.Redact(*value)
	}
	items := make([]map[string]any, 0)
	for hostname, components := range report {
		if hostname != host || components == nil {
			return bad()
		}
		for component, raw := range components {
			var details *struct {
				Name        *string `json:"name"`
				Version     *string `json:"version"`
				ReleaseDate *string `json:"release_date"`
				Status      *struct {
					Health *string `json:"health"`
					State  *string `json:"state"`
				} `json:"status"`
			}
			if component == "" || json.Unmarshal(raw, &details) != nil || details == nil {
				return bad()
			}
			redacted, err := security.RedactJSON(raw)
			if err != nil {
				return bad()
			}
			encoded, err := json.MarshalIndent(redacted, "", "  ")
			if err != nil {
				return bad()
			}
			identity, _ := json.Marshal([]string{hostname, component})
			item := map[string]any{"id": string(identity), "host": hostname, "component": component, "name": text(details.Name), "version": text(details.Version), "release_date": text(details.ReleaseDate), "health": nil, "state": nil, "details": string(encoded)}
			if details.Status != nil {
				item["health"], item["state"] = text(details.Status.Health), text(details.Status.State)
			}
			items = append(items, item)
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i]["id"].(string) < items[j]["id"].(string) })
	return map[string]any{"host": host, "category": "firmwares", "items": items, "observed_at": time.Now().UTC()}, nil
}
