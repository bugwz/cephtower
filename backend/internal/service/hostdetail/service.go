package hostdetail

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	cephdomain "cephtower/backend/internal/domain/ceph"
	"cephtower/backend/internal/integration/ceph/executor"
	"cephtower/backend/internal/security"
	clusterservice "cephtower/backend/internal/service/cluster"
)

type Service struct {
	clusters *clusterservice.Service
	executor executor.Executor
}

func New(clusters *clusterservice.Service, runner executor.Executor) *Service {
	return &Service{clusters: clusters, executor: runner}
}

func (s *Service) Devices(ctx context.Context, clusterID uint64, hostname string) ([]map[string]any, error) {
	hostname = strings.TrimSpace(hostname)
	if clusterID == 0 || hostname == "" {
		return nil, invalid("cluster_id and hostname are required")
	}
	var devices []map[string]any
	if err := s.runJSON(ctx, clusterID, "host.devices", []string{"device", "ls-by-host", hostname, "--format", "json"}, &devices); err != nil {
		return nil, err
	}
	return devices, nil
}

func (s *Service) SMART(ctx context.Context, clusterID uint64, hostname string) (map[string]any, error) {
	devices, err := s.Devices(ctx, clusterID, hostname)
	if err != nil {
		return nil, err
	}
	daemons := make(map[string]struct{})
	for _, device := range devices {
		for _, daemon := range stringValues(device["daemons"]) {
			if strings.HasPrefix(daemon, "mon.") || strings.HasPrefix(daemon, "osd.") {
				daemons[daemon] = struct{}{}
			}
		}
	}
	result := make(map[string]any)
	for daemon := range daemons {
		var payload map[string]any
		err := s.runJSON(ctx, clusterID, "host.smart", []string{"device", "query-daemon-health-metrics", daemon, "--format", "json"}, &payload)
		if err != nil {
			continue
		}
		for deviceID, data := range payload {
			result[deviceID] = data
		}
	}
	return result, nil
}

func (s *Service) runJSON(ctx context.Context, clusterID uint64, id string, args []string, target any) error {
	access, err := s.clusters.Access(ctx, clusterID)
	if err != nil {
		return err
	}
	defer func() { access.ClientKey = "" }()
	result, err := s.executor.Run(ctx, access, executor.CommandSpec{
		ID: id, Binary: executor.BinaryCeph, Args: args,
		Timeout: 30 * time.Second, MaxOutput: executor.DefaultMaxOutput,
	})
	if err != nil {
		return &cephdomain.ActionError{Code: "ceph_command_failed", Message: security.Redact(err.Error()), Retryable: true}
	}
	if err := json.Unmarshal(result.Stdout, target); err != nil {
		return &cephdomain.ActionError{Code: "invalid_ceph_response", Message: "failed to parse Ceph response", Retryable: true}
	}
	return nil
}

func stringValues(value any) []string {
	items, ok := value.([]any)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(items))
	for _, item := range items {
		if text, ok := item.(string); ok && text != "" {
			result = append(result, text)
		}
	}
	return result
}

func invalid(message string) error {
	return &cephdomain.ActionError{Code: "invalid_request", Message: fmt.Sprintf("host detail: %s", message)}
}
