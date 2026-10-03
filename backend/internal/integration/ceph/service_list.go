package ceph

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"cephtower/backend/internal/integration/ceph/executor"
)

func (p *NativeProvider) collectServiceList(ctx context.Context, access ClusterAccess) ([]serviceWire, error) {
	result, err := p.Executor.Run(ctx, access, executor.CommandSpec{ID: "collect.service", Binary: executor.BinaryCeph, Args: []string{"orch", "ls", "--refresh", "--format", "json"}, Timeout: 45 * time.Second, MaxOutput: executor.DefaultMaxOutput})
	if err != nil {
		return nil, err
	}
	// Orchestrator returns this literal before checking the requested format.
	if strings.TrimSpace(string(result.Stdout)) == "No services reported" {
		return []serviceWire{}, nil
	}
	if !json.Valid(result.Stdout) {
		return nil, fmt.Errorf("parse collect.service response: expected a JSON service array")
	}
	decoder := json.NewDecoder(bytes.NewReader(result.Stdout))
	decoder.UseNumber()
	var services []serviceWire
	if err := decoder.Decode(&services); err != nil {
		return nil, fmt.Errorf("parse collect.service response: %w", err)
	}
	if services == nil {
		return nil, fmt.Errorf("parse collect.service response: null is not a service array")
	}
	seen := make(map[string]bool, len(services))
	for _, service := range services {
		if strings.TrimSpace(service.ServiceName) == "" || strings.TrimSpace(service.ServiceType) == "" || seen[service.ServiceName] {
			return nil, fmt.Errorf("parse collect.service response: missing or duplicate service identity")
		}
		seen[service.ServiceName] = true
	}
	return services, nil
}
