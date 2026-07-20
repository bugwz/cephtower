package typed

import (
	"context"
	"encoding/json"
	"net/url"
)

type EmptyResponse struct{}

type DoFunc func(ctx context.Context, method string, pathTemplate string, path map[string]string, query url.Values, body any, auth bool) (json.RawMessage, error)

type Client struct {
	do DoFunc
}

func NewClient(do DoFunc) *Client {
	return &Client{do: do}
}

func (c *Client) doJSON(ctx context.Context, method string, pathTemplate string, path map[string]string, query url.Values, body any, auth bool, out any) error {
	payload, err := c.do(ctx, method, pathTemplate, path, query, body, auth)
	if err != nil {
		return err
	}
	if out == nil || len(payload) == 0 {
		return nil
	}

	return json.Unmarshal(payload, out)
}
