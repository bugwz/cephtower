package ceph

import (
	"context"
	"net/http"
)

func (c *DashboardClient) ClusterSummary(ctx context.Context) (ClusterSummary, error) {
	if c.baseURL == "" {
		return ClusterSummary{HealthStatus: "unknown"}, nil
	}

	var summary ClusterSummary
	if err := c.Do(ctx, Request{
		Method: http.MethodGet,
		Path:   "/api/summary",
		Auth:   true,
	}, &summary); err != nil {
		return ClusterSummary{}, err
	}

	return summary, nil
}

func (c *DashboardClient) Version(ctx context.Context) (string, error) {
	var payload struct {
		Version string `json:"version"`
	}
	if err := c.Do(ctx, Request{
		Method: http.MethodGet,
		Path:   "/api/version",
		Auth:   true,
	}, &payload); err != nil {
		return "", err
	}

	return payload.Version, nil
}

func (c *DashboardClient) HealthFull(ctx context.Context) (map[string]any, error) {
	var payload map[string]any
	if err := c.Do(ctx, Request{
		Method: http.MethodGet,
		Path:   "/api/health/full",
		Auth:   true,
	}, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (c *DashboardClient) HealthMinimal(ctx context.Context) (map[string]any, error) {
	var payload map[string]any
	if err := c.Do(ctx, Request{
		Method: http.MethodGet,
		Path:   "/api/health/minimal",
		Auth:   true,
	}, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}
