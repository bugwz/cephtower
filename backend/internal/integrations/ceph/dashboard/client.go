package dashboard

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"cephtower/backend/internal/integrations/ceph/dashboard/endpoints"
	"cephtower/backend/internal/integrations/ceph/dashboard/typed"
)

const (
	defaultTimeout = 15 * time.Second
	jsonContent    = "application/json"
	cephAPIJSON    = "application/vnd.ceph.api.v1.0+json"
)

var ErrDashboardNotConfigured = errors.New("ceph dashboard base URL is not configured")

type Config struct {
	BaseURL     string
	Username    string
	Password    string
	InsecureTLS bool
}

type DashboardClient struct {
	baseURL  string
	username string
	password string
	client   *http.Client

	authMu sync.Mutex
	token  string
}

type Request struct {
	Method string
	Path   string
	Query  url.Values
	Body   any
	Auth   bool
}

type OperationRequest = endpoints.OperationRequest

type APIError struct {
	Method     string
	URL        string
	StatusCode int
	Status     string
	Body       string
}

func (e *APIError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("ceph dashboard %s %s failed: %s", e.Method, e.URL, e.Status)
	}

	return fmt.Sprintf("ceph dashboard %s %s failed: %s: %s", e.Method, e.URL, e.Status, e.Body)
}

func NewDashboardClient(cfg Config) *DashboardClient {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if cfg.InsecureTLS {
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec
	}

	return &DashboardClient{
		baseURL:  cfg.BaseURL,
		username: strings.TrimSpace(cfg.Username),
		password: cfg.Password,
		client: &http.Client{
			Timeout:   defaultTimeout,
			Transport: transport,
		},
	}
}

func NewDashboardClientWithHTTPClient(cfg Config, httpClient *http.Client) *DashboardClient {
	c := NewDashboardClient(cfg)
	if httpClient != nil {
		c.client = httpClient
	}

	return c
}

func (c *DashboardClient) API() *endpoints.Client {
	return endpoints.NewClient(c.rawOperation)
}

func (c *DashboardClient) TypedAPI() *typed.Client {
	return typed.NewClient(c.typedOperation)
}

func (c *DashboardClient) Do(ctx context.Context, request Request, out any) error {
	if strings.TrimSpace(c.baseURL) == "" {
		return ErrDashboardNotConfigured
	}

	if strings.TrimSpace(request.Method) == "" {
		request.Method = http.MethodGet
	}

	if err := c.do(ctx, request, out); err != nil {
		var apiErr *APIError
		if request.Auth && errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusUnauthorized {
			c.clearToken()
			return c.do(ctx, request, out)
		}

		return err
	}

	return nil
}

func (c *DashboardClient) Raw(ctx context.Context, method string, path string, query url.Values, body any) (json.RawMessage, error) {
	var payload json.RawMessage
	err := c.Do(ctx, Request{
		Method: method,
		Path:   path,
		Query:  query,
		Body:   body,
		Auth:   true,
	}, &payload)
	return payload, err
}

func (c *DashboardClient) rawOperation(ctx context.Context, method string, pathTemplate string, request OperationRequest, auth bool) (json.RawMessage, error) {
	path, err := renderOperationPath(pathTemplate, request.Path)
	if err != nil {
		return nil, err
	}

	var payload json.RawMessage
	err = c.Do(ctx, Request{
		Method: method,
		Path:   path,
		Query:  request.Query,
		Body:   request.Body,
		Auth:   auth,
	}, &payload)
	return payload, err
}

func (c *DashboardClient) typedOperation(ctx context.Context, method string, pathTemplate string, path map[string]string, query url.Values, body any, auth bool) (json.RawMessage, error) {
	return c.rawOperation(ctx, method, pathTemplate, OperationRequest{
		Path:  path,
		Query: query,
		Body:  body,
	}, auth)
}

func (c *DashboardClient) do(ctx context.Context, request Request, out any) error {
	req, err := c.newHTTPRequest(ctx, request)
	if err != nil {
		return err
	}

	if request.Auth {
		token, err := c.authToken(ctx)
		if err != nil {
			return err
		}
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return newAPIError(req, resp)
	}

	if out == nil || resp.StatusCode == http.StatusNoContent {
		io.Copy(io.Discard, resp.Body)
		return nil
	}

	if raw, ok := out.(*json.RawMessage); ok {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		*raw = data
		return nil
	}

	return json.NewDecoder(resp.Body).Decode(out)
}

func (c *DashboardClient) newHTTPRequest(ctx context.Context, request Request) (*http.Request, error) {
	endpoint, err := c.endpoint(request.Path, request.Query)
	if err != nil {
		return nil, err
	}

	var body io.Reader
	if request.Body != nil {
		data, err := json.Marshal(request.Body)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, request.Method, endpoint, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", cephAPIJSON)
	if request.Body != nil {
		req.Header.Set("Content-Type", jsonContent)
	}

	return req, nil
}

func (c *DashboardClient) endpoint(apiPath string, query url.Values) (string, error) {
	if !strings.HasPrefix(apiPath, "/") {
		apiPath = "/" + apiPath
	}

	endpoint, err := url.JoinPath(c.baseURL, apiPath)
	if err != nil {
		return "", err
	}

	if len(query) == 0 {
		return endpoint, nil
	}

	parsed, err := url.Parse(endpoint)
	if err != nil {
		return "", err
	}
	parsed.RawQuery = query.Encode()
	return parsed.String(), nil
}

func renderOperationPath(pathTemplate string, params map[string]string) (string, error) {
	var missing []string
	path := pathTemplateParamRE.ReplaceAllStringFunc(pathTemplate, func(match string) string {
		name := strings.TrimSuffix(strings.TrimPrefix(match, "{"), "}")
		value, ok := params[name]
		if !ok {
			missing = append(missing, name)
			return match
		}

		return url.PathEscape(value)
	})

	if len(missing) > 0 {
		return "", fmt.Errorf("missing ceph dashboard path parameter(s): %s", strings.Join(missing, ", "))
	}

	return path, nil
}

func (c *DashboardClient) clearToken() {
	c.authMu.Lock()
	defer c.authMu.Unlock()

	c.token = ""
}

func newAPIError(req *http.Request, resp *http.Response) error {
	data, err := io.ReadAll(io.LimitReader(resp.Body, 4096))
	if err != nil {
		data = nil
	}

	return &APIError{
		Method:     req.Method,
		URL:        req.URL.Redacted(),
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Body:       strings.TrimSpace(string(data)),
	}
}
