package clusterinspect

import (
	"context"
	"encoding/json"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/security"
)

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
