package ceph

import (
	"context"
	"net/http"
	"net/url"
)

func (c *DashboardClient) ListDaemons(ctx context.Context, daemonTypes string) ([]map[string]any, error) {
	query := url.Values{}
	if daemonTypes != "" {
		query.Set("daemon_types", daemonTypes)
	}

	var payload []map[string]any
	if err := c.Do(ctx, Request{
		Method: http.MethodGet,
		Path:   "/api/daemon",
		Query:  query,
		Auth:   true,
	}, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (c *DashboardClient) ApplyDaemonAction(ctx context.Context, daemonName string, request DaemonActionRequest) error {
	return c.Do(ctx, Request{
		Method: http.MethodPut,
		Path:   "/api/daemon/" + url.PathEscape(daemonName),
		Body:   request,
		Auth:   true,
	}, nil)
}
