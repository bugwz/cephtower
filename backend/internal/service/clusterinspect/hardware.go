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
	if clusterID == 0 || (host != "" && !regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,255}$`).MatchString(host)) {
		return nil, invalid("invalid hardware host or cluster id")
	}
	switch category {
	case "firmwares":
		if host == "" {
			return nil, invalid("firmware inventory requires a host")
		}
		return s.hardwareFirmwares(ctx, clusterID, host)
	case "memory", "storage", "processors", "network", "power", "fans":
	default:
		return nil, invalid("invalid hardware category")
	}
	var report map[string]map[string]map[string]json.RawMessage
	args := []string{"orch", "hardware", "status"}
	if host != "" {
		args = append(args, "--hostname", host)
	}
	args = append(args, "--category", category, "--format", "json")
	if err := s.read(ctx, clusterID, "host.hardware", args, &report); err != nil {
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
		if hostname == "" || (host != "" && hostname != host) || systems == nil {
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
				identity, _ := json.Marshal([]string{hostname, system, component})
				redacted, err := security.RedactJSON(raw)
				if err != nil {
					return bad()
				}
				encoded, err := json.MarshalIndent(redacted, "", "  ")
				if err != nil {
					return bad()
				}
				items = append(items, map[string]any{"id": string(identity), "host": hostname, "system": system, "component": component, "health": health, "state": state, "details": string(encoded)})
			}
		}
	}
	sort.Slice(items, func(i, j int) bool { return items[i]["id"].(string) < items[j]["id"].(string) })
	return map[string]any{"host": host, "category": category, "items": items, "observed_at": time.Now().UTC()}, nil
}
