package ceph

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

func (c *DashboardClient) ListHosts(ctx context.Context, options ListHostsOptions) ([]Host, error) {
	var hosts []Host
	if err := c.Do(ctx, Request{
		Method: http.MethodGet,
		Path:   "/api/host",
		Query:  listHostsQuery(options),
		Auth:   true,
	}, &hosts); err != nil {
		return nil, err
	}

	return hosts, nil
}

func (c *DashboardClient) HostDetails(ctx context.Context, hostname string) (map[string]any, error) {
	var payload map[string]any
	if err := c.Do(ctx, Request{
		Method: http.MethodGet,
		Path:   "/api/host/" + url.PathEscape(hostname),
		Auth:   true,
	}, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (c *DashboardClient) CreateHost(ctx context.Context, request HostRequest) error {
	return c.Do(ctx, Request{
		Method: http.MethodPost,
		Path:   "/api/host",
		Body:   request,
		Auth:   true,
	}, nil)
}

func (c *DashboardClient) UpdateHost(ctx context.Context, hostname string, request UpdateHostRequest) error {
	return c.Do(ctx, Request{
		Method: http.MethodPut,
		Path:   "/api/host/" + url.PathEscape(hostname),
		Body:   request,
		Auth:   true,
	}, nil)
}

func (c *DashboardClient) DeleteHost(ctx context.Context, hostname string) error {
	return c.Do(ctx, Request{
		Method: http.MethodDelete,
		Path:   "/api/host/" + url.PathEscape(hostname),
		Auth:   true,
	}, nil)
}

func (c *DashboardClient) HostDaemons(ctx context.Context, hostname string) ([]map[string]any, error) {
	var payload []map[string]any
	if err := c.Do(ctx, Request{
		Method: http.MethodGet,
		Path:   "/api/host/" + url.PathEscape(hostname) + "/daemons",
		Auth:   true,
	}, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (c *DashboardClient) HostDevices(ctx context.Context, hostname string) ([]map[string]any, error) {
	var payload []map[string]any
	if err := c.Do(ctx, Request{
		Method: http.MethodGet,
		Path:   "/api/host/" + url.PathEscape(hostname) + "/devices",
		Auth:   true,
	}, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (c *DashboardClient) HostInventory(ctx context.Context, hostname string) (map[string]any, error) {
	var payload map[string]any
	if err := c.Do(ctx, Request{
		Method: http.MethodGet,
		Path:   "/api/host/" + url.PathEscape(hostname) + "/inventory",
		Auth:   true,
	}, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func listHostsQuery(options ListHostsOptions) url.Values {
	query := url.Values{}
	if options.Sources != "" {
		query.Set("sources", options.Sources)
	}
	if options.Facts != nil {
		query.Set("facts", strconv.FormatBool(*options.Facts))
	}
	if options.Offset != nil {
		query.Set("offset", strconv.Itoa(*options.Offset))
	}
	if options.Limit != nil {
		query.Set("limit", strconv.Itoa(*options.Limit))
	}
	if options.Search != "" {
		query.Set("search", options.Search)
	}
	if options.Sort != "" {
		query.Set("sort", options.Sort)
	}
	if options.IncludeServiceInstances != nil {
		query.Set("include_service_instances", strconv.FormatBool(*options.IncludeServiceInstances))
	}

	return query
}
