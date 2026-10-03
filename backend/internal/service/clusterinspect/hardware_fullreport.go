package clusterinspect

import (
	"context"
	"encoding/json"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/security"
)

func (s *Service) hardwareFullReport(ctx context.Context, clusterID uint64, host string) (map[string]any, error) {
	var report map[string]json.RawMessage
	if err := s.read(ctx, clusterID, "host.hardware", []string{"orch", "hardware", "status", "--hostname", host, "--category", "fullreport", "--format", "json"}, &report); err != nil {
		return nil, err
	}
	bad := func() (map[string]any, error) {
		return nil, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "invalid full hardware report or host identity"}
	}
	var identity struct {
		Host   string  `json:"host"`
		Serial *string `json:"sn"`
	}
	if len(report) != 1 || json.Unmarshal(report[host], &identity) != nil || identity.Host != host {
		return bad()
	}
	redacted, err := security.RedactJSON(report[host])
	if err != nil {
		return bad()
	}
	encoded, err := json.MarshalIndent(redacted, "", "  ")
	if err != nil {
		return bad()
	}
	var serial any
	if identity.Serial != nil {
		serial = security.Redact(*identity.Serial)
	}
	return map[string]any{"host": host, "category": "fullreport", "serial_number": serial, "report": string(encoded), "observed_at": time.Now().UTC()}, nil
}
