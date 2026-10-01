package clusterinspect

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/security"
)

type TelemetryReport struct {
	Mode       string    `json:"mode"`
	ReportJSON string    `json:"report_json,omitempty"`
	Message    string    `json:"message,omitempty"`
	ObservedAt time.Time `json:"observed_at"`
}

func (s *Service) TelemetryReport(ctx context.Context, clusterID uint64, mode string) (TelemetryReport, error) {
	if clusterID == 0 || !oneOf(mode, "current", "preview") {
		return TelemetryReport{}, invalid("cluster_id and telemetry report mode (current or preview) are required")
	}
	access, err := s.clusters.Access(ctx, clusterID)
	if err != nil {
		return TelemetryReport{}, err
	}
	command := "show-all"
	if mode == "preview" {
		command = "preview-all"
	}
	result, err := s.executor.Run(ctx, access, executor.CommandSpec{ID: "telemetry.report", Binary: executor.BinaryCeph, Args: []string{"telemetry", command, "--format", "json"}, Timeout: time.Minute, MaxOutput: executor.DefaultMaxOutput})
	if err != nil {
		return TelemetryReport{}, &cephdomain.ActionError{Code: "ceph_command_failed", Message: security.Redact(err.Error()), Retryable: true}
	}
	return parseTelemetryReport(result.Stdout, mode)
}

func parseTelemetryReport(data []byte, mode string) (TelemetryReport, error) {
	report := TelemetryReport{Mode: mode, ObservedAt: time.Now().UTC()}
	text := strings.TrimSpace(string(data))
	if (mode == "current" && text == "Telemetry is off. Please consider opting-in with `ceph telemetry on`.\nPreview sample reports with `ceph telemetry preview`.") || (mode == "preview" && text == "Telemetry is up to date, see report with `ceph telemetry show`.") {
		report.Message = text
		return report, nil
	}
	var value map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if decoder.Decode(&value) != nil || decoder.Decode(new(any)) != io.EOF || value == nil {
		return TelemetryReport{}, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "Ceph returned neither a telemetry report object nor a recognized availability message"}
	}
	redacted, err := security.RedactJSON(value)
	if err != nil {
		return TelemetryReport{}, err
	}
	encoded, err := json.MarshalIndent(redacted, "", "  ")
	if err != nil {
		return TelemetryReport{}, err
	}
	// Carry JSON as text so the browser cannot round large report counters.
	report.ReportJSON = string(encoded)
	return report, nil
}

func (s *Service) TelemetryStatus(ctx context.Context, clusterID uint64) (map[string]any, error) {
	var status map[string]any
	if err := s.read(ctx, clusterID, "telemetry.status", []string{"telemetry", "status", "--format", "json"}, &status); err != nil {
		return nil, err
	}
	invalidResponse := func() (map[string]any, error) {
		return nil, &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "Ceph returned invalid telemetry status"}
	}
	if _, ok := status["enabled"].(bool); !ok {
		return invalidResponse()
	}
	for _, key := range []string{"channel_basic", "channel_ident", "channel_crash", "channel_device", "channel_perf", "leaderboard"} {
		if value, present := status[key]; present {
			if _, ok := value.(bool); !ok {
				return invalidResponse()
			}
		}
	}
	for _, key := range []string{"interval", "last_opt_revision"} {
		if value, present := status[key]; present {
			number, ok := value.(json.Number)
			if !ok {
				return invalidResponse()
			}
			integer, err := number.Int64()
			if err != nil || integer < 0 {
				return invalidResponse()
			}
			status[key] = number.String()
		}
	}
	for _, key := range []string{"url", "device_url", "proxy", "contact", "organization", "description", "leaderboard_description", "last_upload"} {
		if value, present := status[key]; present && value != nil {
			if number, ok := value.(json.Number); key == "last_upload" && ok && number.String() == "0" {
				continue
			}
			if _, ok := value.(string); !ok {
				return invalidResponse()
			}
		}
	}
	redacted, err := security.RedactJSON(status)
	if err != nil {
		return nil, err
	}
	return map[string]any{"status": redacted, "observed_at": time.Now().UTC()}, nil
}
