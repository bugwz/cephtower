package monitoring

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const maxResponseBytes = 4 << 20

type Client struct {
	baseURL *url.URL
	http    *http.Client
	token   string
}

func New(rawURL, token string, client *http.Client) (*Client, error) {
	base, err := url.Parse(strings.TrimSpace(rawURL))
	if err != nil || base.Host == "" || (base.Scheme != "http" && base.Scheme != "https") || base.User != nil {
		return nil, fmt.Errorf("endpoint must be an http(s) URL without embedded credentials")
	}
	if client == nil {
		client = &http.Client{Timeout: 15 * time.Second}
	}
	return &Client{baseURL: base, http: client, token: strings.TrimSpace(token)}, nil
}

type PrometheusResult struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string            `json:"resultType"`
		Result     []json.RawMessage `json:"result"`
	} `json:"data"`
}

var metricQueries = map[string]string{
	"cluster_health":        "ceph_health_status",
	"capacity_used_percent": "100 * (ceph_cluster_total_used_bytes / ceph_cluster_total_bytes)",
	"client_read_bytes":     "sum(rate(ceph_pool_rd_bytes[5m]))",
	"client_write_bytes":    "sum(rate(ceph_pool_wr_bytes[5m]))",
}

func (c *Client) Query(ctx context.Context, metricID string, at *time.Time) (PrometheusResult, error) {
	query, ok := metricQueries[metricID]
	if !ok {
		return PrometheusResult{}, fmt.Errorf("unsupported metric_id %q", metricID)
	}
	values := url.Values{"query": []string{query}}
	if at != nil {
		values.Set("time", at.UTC().Format(time.RFC3339Nano))
	}
	var result PrometheusResult
	err := c.get(ctx, "/api/v1/query?"+values.Encode(), &result)
	return result, err
}
func (c *Client) QueryRange(ctx context.Context, metricID string, start, end time.Time, step time.Duration) (PrometheusResult, error) {
	query, ok := metricQueries[metricID]
	if !ok {
		return PrometheusResult{}, fmt.Errorf("unsupported metric_id %q", metricID)
	}
	if !end.After(start) || end.Sub(start) > 31*24*time.Hour || step < time.Second {
		return PrometheusResult{}, fmt.Errorf("invalid query range")
	}
	values := url.Values{"query": []string{query}, "start": []string{start.UTC().Format(time.RFC3339Nano)}, "end": []string{end.UTC().Format(time.RFC3339Nano)}, "step": []string{strconv.FormatFloat(step.Seconds(), 'f', -1, 64)}}
	var result PrometheusResult
	err := c.get(ctx, "/api/v1/query_range?"+values.Encode(), &result)
	return result, err
}

type Alert struct {
	Labels      map[string]string `json:"labels"`
	Annotations map[string]string `json:"annotations"`
	Status      struct {
		State string `json:"state"`
	} `json:"status"`
	StartsAt time.Time `json:"startsAt"`
}
type RuleGroup struct {
	Name  string `json:"name"`
	File  string `json:"file,omitempty"`
	Rules []Rule `json:"rules"`
}
type Rule struct {
	Name        string            `json:"name"`
	Query       string            `json:"query"`
	Duration    float64           `json:"duration,omitempty"`
	Labels      map[string]string `json:"labels,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
	State       string            `json:"state,omitempty"`
	Health      string            `json:"health,omitempty"`
}
type Silence struct {
	ID        string    `json:"id,omitempty"`
	Matchers  []Matcher `json:"matchers"`
	StartsAt  time.Time `json:"startsAt"`
	EndsAt    time.Time `json:"endsAt"`
	CreatedBy string    `json:"createdBy"`
	Comment   string    `json:"comment"`
}
type Matcher struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	IsRegex bool   `json:"isRegex"`
	IsEqual bool   `json:"isEqual"`
}

func (c *Client) Alerts(ctx context.Context) ([]Alert, error) {
	var result []Alert
	err := c.get(ctx, "/api/v2/alerts", &result)
	return result, err
}
func (c *Client) Rules(ctx context.Context) ([]RuleGroup, error) {
	var response struct {
		Status string `json:"status"`
		Data   struct {
			Groups []RuleGroup `json:"groups"`
		} `json:"data"`
	}
	if err := c.get(ctx, "/api/v1/rules", &response); err != nil {
		return nil, err
	}
	if response.Status != "" && response.Status != "success" {
		return nil, fmt.Errorf("Prometheus rules query failed")
	}
	return response.Data.Groups, nil
}
func (c *Client) Silences(ctx context.Context) ([]Silence, error) {
	var result []Silence
	err := c.get(ctx, "/api/v2/silences", &result)
	return result, err
}
func (c *Client) CreateSilence(ctx context.Context, silence Silence) (string, error) {
	if !silence.EndsAt.After(silence.StartsAt) || len(silence.Matchers) == 0 {
		return "", fmt.Errorf("silence requires matchers and a valid time range")
	}
	var result struct {
		SilenceID string `json:"silenceID"`
	}
	if err := c.jsonRequest(ctx, http.MethodPost, "/api/v2/silences", silence, &result); err != nil {
		return "", err
	}
	return result.SilenceID, nil
}
func (c *Client) DeleteSilence(ctx context.Context, id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("silence id is required")
	}
	return c.jsonRequest(ctx, http.MethodDelete, "/api/v2/silence/"+url.PathEscape(id), nil, nil)
}

type Dashboard struct {
	ID    string `json:"id"`
	UID   string `json:"uid"`
	Title string `json:"title"`
	URL   string `json:"url"`
}

func (c *Client) Dashboards(ctx context.Context) ([]Dashboard, error) {
	var result []Dashboard
	err := c.get(ctx, "/api/search?type=dash-db", &result)
	return result, err
}

func (c *Client) get(ctx context.Context, path string, out any) error {
	return c.jsonRequest(ctx, http.MethodGet, path, nil, out)
}
func (c *Client) jsonRequest(ctx context.Context, method, path string, body, out any) error {
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
	if c.token != "" {
		request.Header.Set("Authorization", "Bearer "+c.token)
	}
	response, err := c.http.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	limited := io.LimitReader(response.Body, maxResponseBytes+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return err
	}
	if len(data) > maxResponseBytes {
		return fmt.Errorf("endpoint response exceeds %d bytes", maxResponseBytes)
	}
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return fmt.Errorf("endpoint returned HTTP %d", response.StatusCode)
	}
	if out == nil || len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("decode endpoint response: %w", err)
	}
	return nil
}
