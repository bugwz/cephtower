package clusterinspect

import (
	"context"
	"encoding/json"
	"sort"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/security"
)

func (s *Service) hardwareCriticals(ctx context.Context, clusterID uint64, host string) (map[string]any, error) {
	var report map[string]map[string]map[string]map[string]json.RawMessage
	args := []string{"orch", "hardware", "status"}
	if host != "" {
		args = append(args, "--hostname", host)
	}
	args = append(args, "--category", "criticals", "--format", "json")
	if err := s.read(ctx, clusterID, "host.hardware", args, &report); err != nil {
		return nil, err
	}
	bad := func() (map[string]any, error) {
		return nil, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "invalid hardware non-OK report"}
	}
	if report == nil {
		return bad()
	}
	items := make([]map[string]any, 0)
	for hostname, systems := range report {
		if hostname == "" || (host != "" && hostname != host) || systems == nil {
			return bad()
		}
		for system, categories := range systems {
			if system == "" || categories == nil {
				return bad()
			}
			for category, members := range categories {
				if category == "" || members == nil {
					return bad()
				}
				for component, raw := range members {
					var details *struct {
						Description string `json:"description"`
						Name        string `json:"name"`
						ID          string `json:"id"`
						Status      *struct {
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
					name := details.Description
					if name == "" {
						name = details.Name
					}
					if name == "" {
						name = details.ID
					}
					redacted, err := security.RedactJSON(raw)
					if err != nil {
						return bad()
					}
					encoded, err := json.MarshalIndent(redacted, "", "  ")
					if err != nil {
						return bad()
					}
					identity, _ := json.Marshal([]string{hostname, system, category, component})
					items = append(items, map[string]any{"id": string(identity), "host": hostname, "system": system, "category": category, "component": component, "name": security.Redact(name), "health": health, "state": state, "details": string(encoded)})
				}
			}
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i]["id"].(string) < items[j]["id"].(string) })
	return map[string]any{"host": host, "category": "criticals", "items": items, "observed_at": time.Now().UTC()}, nil
}
