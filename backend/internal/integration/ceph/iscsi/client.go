package iscsi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const maxResponseBytes = 4 << 20

type Client struct {
	baseURL            *url.URL
	http               *http.Client
	username, password string
}
type GatewayHealth struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}
type Target struct {
	IQN     string      `json:"iqn"`
	Portals []Portal    `json:"portals"`
	Disks   []Disk      `json:"disks"`
	Clients []Initiator `json:"clients"`
	Groups  []Group     `json:"groups"`
}
type Portal struct {
	Host string `json:"host"`
	IP   string `json:"ip"`
}
type Disk struct {
	Pool      string `json:"pool"`
	Image     string `json:"image"`
	Backstore string `json:"backstore"`
}
type Initiator struct {
	IQN  string   `json:"iqn"`
	LUNs []string `json:"luns"`
}
type Group struct {
	Name    string   `json:"name"`
	Members []string `json:"members"`
	Disks   []string `json:"disks"`
}

func New(rawURL, username, password string, client *http.Client) (*Client, error) {
	base, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || base.Host == "" || base.Scheme != "https" || base.User != nil {
		return nil, fmt.Errorf("iSCSI endpoint must be an https URL without embedded credentials")
	}
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	return &Client{baseURL: base, http: client, username: username, password: password}, nil
}
func (c *Client) Health(ctx context.Context) (GatewayHealth, error) {
	var result GatewayHealth
	err := c.request(ctx, http.MethodGet, "/api/gateway", nil, &result)
	return result, err
}
func (c *Client) Targets(ctx context.Context) ([]Target, error) {
	var result []Target
	err := c.request(ctx, http.MethodGet, "/api/target", nil, &result)
	return result, err
}
func (c *Client) Target(ctx context.Context, iqn string) (Target, error) {
	var result Target
	err := c.request(ctx, http.MethodGet, "/api/target/"+url.PathEscape(iqn), nil, &result)
	return result, err
}
func (c *Client) ApplyTarget(ctx context.Context, target Target) error {
	if strings.TrimSpace(target.IQN) == "" {
		return fmt.Errorf("target iqn is required")
	}
	return c.request(ctx, http.MethodPut, "/api/target/"+url.PathEscape(target.IQN), target, nil)
}
func (c *Client) DeleteTarget(ctx context.Context, iqn string) error {
	if strings.TrimSpace(iqn) == "" {
		return fmt.Errorf("target iqn is required")
	}
	return c.request(ctx, http.MethodDelete, "/api/target/"+url.PathEscape(iqn), nil, nil)
}
func (c *Client) request(ctx context.Context, method, path string, body, out any) error {
	endpoint, err := c.baseURL.Parse(path)
	if err != nil {
		return err
	}
	var reader io.Reader
	if body != nil {
		encoded, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reader = strings.NewReader(string(encoded))
	}
	request, err := http.NewRequestWithContext(ctx, method, endpoint.String(), reader)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if c.username != "" {
		request.SetBasicAuth(c.username, c.password)
	}
	response, err := c.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	data, err := io.ReadAll(io.LimitReader(response.Body, maxResponseBytes+1))
	if err != nil {
		return err
	}
	if len(data) > maxResponseBytes {
		return fmt.Errorf("iSCSI response exceeds %d bytes", maxResponseBytes)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("iSCSI gateway returned HTTP %d", response.StatusCode)
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}
