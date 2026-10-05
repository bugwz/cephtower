package clusterinspect

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"
	"unicode"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/security"
)

type RGWDaemonStatus struct {
	ServiceMapID string         `json:"service_map_id"`
	StatusStamp  string         `json:"status_stamp"`
	LastBeacon   string         `json:"last_beacon"`
	Status       map[string]any `json:"status"`
	ObservedAt   time.Time      `json:"observed_at"`
}

func (s *Service) RGWDaemonStatus(ctx context.Context, clusterID uint64, id string) (RGWDaemonStatus, error) {
	var empty RGWDaemonStatus
	if clusterID == 0 || id == "" || len(id) > 256 || strings.ContainsFunc(id, unicode.IsControl) {
		return empty, invalid("cluster_id and service_map_id are required")
	}
	access, err := s.clusters.Access(ctx, clusterID)
	if err != nil {
		return empty, err
	}
	if err = ctx.Err(); err != nil {
		return empty, err
	}
	result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: "rgw.daemon.status", Binary: executor.BinaryCeph, Args: []string{"service", "status", "--format", "json"}, Timeout: 20 * time.Second, MaxOutput: executor.DefaultMaxOutput})
	defer clear(result.Stdout)
	defer clear(result.Stderr)
	if ctx.Err() != nil {
		return empty, ctx.Err()
	}
	if err != nil || result.ExitCode != 0 {
		return empty, &cephdomain.ActionError{Code: "ceph_command_failed", Message: "RGW service status could not be read"}
	}
	return decodeRGWDaemonStatus(result.Stdout, id)
}

func decodeRGWDaemonStatus(body []byte, id string) (RGWDaemonStatus, error) {
	var empty RGWDaemonStatus
	fail := func() (RGWDaemonStatus, error) {
		return empty, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "RGW service status is invalid"}
	}
	var root map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(body))
	if decoder.Decode(&root) != nil || decoder.Decode(new(any)) != io.EOF || root == nil {
		return fail()
	}
	missing := func() (RGWDaemonStatus, error) {
		return empty, &cephdomain.ActionError{Code: "resource_not_found", Message: "RGW registration has no service status"}
	}
	raw, exists := root["rgw"]
	if !exists {
		return missing()
	}
	var daemons map[string]json.RawMessage
	if json.Unmarshal(raw, &daemons) != nil || daemons == nil {
		return fail()
	}
	raw, exists = daemons[id]
	if !exists {
		return missing()
	}
	var record struct {
		StatusStamp *string           `json:"status_stamp"`
		LastBeacon  *string           `json:"last_beacon"`
		Status      map[string]string `json:"status"`
	}
	if json.Unmarshal(raw, &record) != nil || record.StatusStamp == nil || record.LastBeacon == nil || record.Status == nil {
		return fail()
	}
	safe := map[string]any{}
	for key, value := range record.Status {
		if security.IsSensitiveName(key) {
			safe[key] = "[REDACTED]"
			continue
		}
		if key == "json" {
			decoded, err := security.RedactJSON(json.RawMessage(value))
			if err != nil {
				return fail()
			}
			safe[key] = decoded
		} else {
			safe[key] = security.Redact(value)
		}
	}
	return RGWDaemonStatus{ServiceMapID: id, StatusStamp: *record.StatusStamp, LastBeacon: *record.LastBeacon, Status: safe, ObservedAt: time.Now().UTC()}, nil
}
