package ceph

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"cephtower/backend/internal/config"
)

type DashboardClient struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

type ClusterSummary struct {
	HealthStatus string `json:"health_status"`
	Version      string `json:"version,omitempty"`
}

func NewDashboardClient(cfg config.CephDashboardConfig) *DashboardClient {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.InsecureTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}

	return &DashboardClient{
		baseURL:  cfg.BaseURL,
		username: cfg.Username,
		password: cfg.Password,
		client: &http.Client{
			Timeout:   15 * time.Second,
			Transport: transport,
		},
	}
}

func (c *DashboardClient) ClusterSummary(ctx context.Context) (ClusterSummary, error) {
	if c.baseURL == "" {
		return ClusterSummary{HealthStatus: "unknown"}, nil
	}

	version, err := c.getVersion(ctx)
	if err != nil {
		return ClusterSummary{}, err
	}

	return ClusterSummary{
		HealthStatus: "unknown",
		Version:      version,
	}, nil
}

func (c *DashboardClient) getVersion(ctx context.Context) (string, error) {
	endpoint, err := url.JoinPath(c.baseURL, "/api/version")
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}

	token, err := c.login(ctx)
	if err != nil {
		return "", err
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("ceph dashboard version request failed: %s", resp.Status)
	}

	var payload struct {
		Version string `json:"version"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}

	return payload.Version, nil
}

func (c *DashboardClient) login(ctx context.Context) (string, error) {
	if c.username == "" || c.password == "" {
		return "", nil
	}

	endpoint, err := url.JoinPath(c.baseURL, "/api/auth")
	if err != nil {
		return "", err
	}

	body, err := json.Marshal(map[string]string{
		"username": c.username,
		"password": c.password,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("ceph dashboard auth failed: %s", resp.Status)
	}

	var payload struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return "", err
	}

	return payload.Token, nil
}
