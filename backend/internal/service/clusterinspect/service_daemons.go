package clusterinspect

import (
	"context"
	"encoding/json"
	"regexp"
	"strconv"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/security"
)

func (s *Service) ServiceDaemons(ctx context.Context, clusterID uint64, name string) (map[string]any, error) {
	if clusterID == 0 || !regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9_.-]{0,255}$`).MatchString(name) {
		return nil, invalid("invalid service name or cluster id")
	}
	var rows []map[string]any
	if err := s.read(ctx, clusterID, "service.daemons", []string{"orch", "ps", "--service-name", name, "--refresh", "--format", "json"}, &rows); err != nil {
		return nil, err
	}
	bad := func() (map[string]any, error) {
		return nil, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "invalid service daemon response"}
	}
	if rows == nil {
		return bad()
	}
	seen := map[string]bool{}
	for _, row := range rows {
		daemon, ok := row["daemon_name"].(string)
		if !ok || daemon == "" || seen[daemon] {
			return bad()
		}
		seen[daemon] = true
		if service, exists := row["service_name"]; exists && service != name {
			return bad()
		}
		// Rank metadata is integer-valued and must not lose precision in JavaScript.
		for _, key := range []string{"rank", "rank_generation"} {
			if value, exists := row[key]; exists && value != nil {
				number, ok := value.(json.Number)
				if !ok || !regexp.MustCompile(`^-?(0|[1-9][0-9]*)$`).MatchString(number.String()) {
					return bad()
				}
				row[key] = number.String()
			}
		}
		// Keep exact byte counts when JavaScript receives the response.
		for _, key := range []string{"memory_usage", "memory_request", "memory_limit"} {
			if value, exists := row[key]; exists && value != nil {
				number, ok := value.(json.Number)
				if !ok {
					return bad()
				}
				if _, err := strconv.ParseUint(number.String(), 10, 64); err != nil {
					return bad()
				}
				row[key] = number.String()
			}
		}
	}
	redacted, err := security.RedactJSON(rows)
	if err != nil {
		return nil, err
	}
	return map[string]any{"items": redacted, "service_name": name, "observed_at": time.Now().UTC()}, nil
}
