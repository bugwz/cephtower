package ceph

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

type ListOSDsOptions struct {
	Offset *int
	Limit  *int
	Search string
	Sort   string
}

func (c *DashboardClient) ListOSDs(ctx context.Context, options ListOSDsOptions) ([]map[string]any, error) {
	var payload []map[string]any
	if err := c.Do(ctx, Request{
		Method: http.MethodGet,
		Path:   "/api/osd",
		Query:  listOSDsQuery(options),
		Auth:   true,
	}, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (c *DashboardClient) GetOSD(ctx context.Context, serviceID string) (map[string]any, error) {
	var payload map[string]any
	if err := c.Do(ctx, Request{
		Method: http.MethodGet,
		Path:   "/api/osd/" + url.PathEscape(serviceID),
		Auth:   true,
	}, &payload); err != nil {
		return nil, err
	}

	return payload, nil
}

func (c *DashboardClient) OSDFlags(ctx context.Context) ([]string, error) {
	var payload struct {
		Flags []string `json:"list_of_flags"`
	}
	if err := c.Do(ctx, Request{
		Method: http.MethodGet,
		Path:   "/api/osd/flags",
		Auth:   true,
	}, &payload); err != nil {
		return nil, err
	}

	return payload.Flags, nil
}

func listOSDsQuery(options ListOSDsOptions) url.Values {
	query := url.Values{}
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

	return query
}
