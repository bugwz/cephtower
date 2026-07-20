package ceph

import (
	"context"
	"net/http"
)

type AuthResponse struct {
	Token             string              `json:"token"`
	Username          string              `json:"username"`
	Permissions       map[string][]string `json:"permissions"`
	PwdExpirationDate string              `json:"pwdExpirationDate"`
	PwdUpdateRequired bool                `json:"pwdUpdateRequired"`
	SSO               bool                `json:"sso"`
}

func (c *DashboardClient) Login(ctx context.Context) (AuthResponse, error) {
	var payload AuthResponse
	if c.username == "" || c.password == "" {
		return payload, nil
	}

	err := c.Do(ctx, Request{
		Method: http.MethodPost,
		Path:   "/api/auth",
		Body: map[string]string{
			"username": c.username,
			"password": c.password,
		},
		Auth: false,
	}, &payload)
	return payload, err
}

func (c *DashboardClient) authToken(ctx context.Context) (string, error) {
	if c.username == "" || c.password == "" {
		return "", nil
	}

	c.authMu.Lock()
	defer c.authMu.Unlock()

	if c.token != "" {
		return c.token, nil
	}

	payload, err := c.Login(ctx)
	if err != nil {
		return "", err
	}

	c.token = payload.Token
	return c.token, nil
}
