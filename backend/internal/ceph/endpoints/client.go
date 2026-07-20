package endpoints

import (
	"context"
	"encoding/json"
	"net/url"
)

type OperationRequest struct {
	Path  map[string]string
	Query url.Values
	Body  any
}

type DoFunc func(ctx context.Context, method string, pathTemplate string, request OperationRequest, auth bool) (json.RawMessage, error)

type Client struct {
	do DoFunc
}

func NewClient(do DoFunc) *Client {
	return &Client{do: do}
}
