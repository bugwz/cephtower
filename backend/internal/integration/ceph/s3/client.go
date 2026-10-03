package s3

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

type Credentials struct {
	AccessKey, SecretKey, SessionToken, Region string
}

type Client struct {
	base        *url.URL
	credentials Credentials
	http        *http.Client
	now         func() time.Time
}

func New(rawURL string, credentials Credentials, client *http.Client) (*Client, error) {
	base, err := url.Parse(rawURL)
	if err != nil || base.Host == "" || (base.Scheme != "http" && base.Scheme != "https") || base.User != nil {
		return nil, fmt.Errorf("invalid S3 endpoint URL")
	}
	if strings.TrimSpace(credentials.AccessKey) == "" || strings.TrimSpace(credentials.SecretKey) == "" {
		return nil, fmt.Errorf("S3 access_key and secret_key are required")
	}
	if credentials.Region == "" {
		credentials.Region = "us-east-1"
	}
	if client == nil {
		client = &http.Client{Timeout: 20 * time.Second}
	}
	return &Client{base: base, credentials: credentials, http: client, now: time.Now}, nil
}

func (c *Client) CreateBucket(ctx context.Context, bucket string) error {
	_, _, err := c.request(ctx, http.MethodPut, bucket, nil, nil)
	return err
}

func (c *Client) DeleteBucket(ctx context.Context, bucket string) error {
	_, _, err := c.request(ctx, http.MethodDelete, bucket, nil, nil)
	return err
}

func (c *Client) HeadBucket(ctx context.Context, bucket string) error {
	_, _, err := c.request(ctx, http.MethodHead, bucket, nil, nil)
	return err
}

func (c *Client) PutBucketConfiguration(ctx context.Context, bucket, kind string, body []byte) error {
	if err := ValidateBucketConfiguration(kind, body); err != nil {
		return err
	}
	_, _, err := c.request(ctx, http.MethodPut, bucket, url.Values{kind: []string{""}}, body)
	return err
}

func (c *Client) GetBucketConfiguration(ctx context.Context, bucket, kind string) ([]byte, string, error) {
	allowed := map[string]bool{"policy": true, "cors": true, "lifecycle": true, "encryption": true, "versioning": true, "tagging": true, "object-lock": true, "acl": true, "replication": true}
	if !allowed[kind] {
		return nil, "", fmt.Errorf("unsupported S3 bucket configuration %q", kind)
	}
	body, contentType, err := c.request(ctx, http.MethodGet, bucket, url.Values{kind: []string{""}}, nil)
	if err == nil && kind == "tagging" {
		err = ValidateBucketConfiguration(kind, body)
	}
	return body, contentType, err
}

func DeletableBucketConfiguration(kind string) bool {
	return kind == "policy" || kind == "cors" || kind == "lifecycle" || kind == "encryption" || kind == "tagging" || kind == "replication"
}

func (c *Client) DeleteBucketConfiguration(ctx context.Context, bucket, kind string) error {
	if !DeletableBucketConfiguration(kind) {
		return fmt.Errorf("unsupported S3 bucket configuration deletion %q", kind)
	}
	_, _, err := c.request(ctx, http.MethodDelete, bucket, url.Values{kind: {""}}, nil)
	return err
}

func (c *Client) request(ctx context.Context, method, bucket string, query url.Values, body []byte) ([]byte, string, error) {
	return c.requestWithHeaders(ctx, method, bucket, query, body, nil)
}

func (c *Client) requestWithHeaders(ctx context.Context, method, bucket string, query url.Values, body []byte, headers http.Header) ([]byte, string, error) {
	if bucket == "" || strings.ContainsAny(bucket, "/\x00") {
		return nil, "", fmt.Errorf("invalid S3 bucket name")
	}
	return c.requestTarget(ctx, method, bucket, query, body, headers)
}

// Only service-level operations may call this with an empty bucket.
func (c *Client) requestTarget(ctx context.Context, method, bucket string, query url.Values, body []byte, headers http.Header) ([]byte, string, error) {
	target := *c.base
	target.Path = strings.TrimSuffix(c.base.Path, "/") + "/" + bucket
	target.RawQuery = canonicalQuery(query)
	req, err := http.NewRequestWithContext(ctx, method, target.String(), bytes.NewReader(body))
	if err != nil {
		return nil, "", err
	}
	now := c.now().UTC()
	for name, values := range headers {
		for _, value := range values {
			req.Header.Add(name, value)
		}
	}
	payloadHash := sha256Hex(body)
	req.Header.Set("X-Amz-Date", now.Format("20060102T150405Z"))
	req.Header.Set("X-Amz-Content-Sha256", payloadHash)
	if len(body) > 0 {
		contentType := "application/xml"
		if query.Has("policy") {
			contentType = "application/json"
		}
		req.Header.Set("Content-Type", contentType)
		checksum := md5.Sum(body) // S3 transport integrity header, not a security hash.
		req.Header.Set("Content-MD5", base64.StdEncoding.EncodeToString(checksum[:]))
	}
	if c.credentials.SessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", c.credentials.SessionToken)
	}
	c.sign(req, now, payloadHash)
	response, err := c.http.Do(req)
	if err != nil {
		return nil, "", err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return nil, "", responseError(response.StatusCode, response.Body)
	}
	data, err := io.ReadAll(io.LimitReader(response.Body, (4<<20)+1))
	if err != nil {
		return nil, "", err
	}
	if len(data) > 4<<20 {
		return nil, "", fmt.Errorf("S3 response exceeds %d bytes", 4<<20)
	}
	return data, response.Header.Get("Content-Type"), nil
}

func (c *Client) sign(req *http.Request, now time.Time, payloadHash string) {
	headerNames := []string{"host", "x-amz-content-sha256", "x-amz-date"}
	if req.Header.Get("X-Amz-Acl") != "" {
		headerNames = append(headerNames, "x-amz-acl")
	}
	if c.credentials.SessionToken != "" {
		headerNames = append(headerNames, "x-amz-security-token")
	}
	sort.Strings(headerNames)
	var canonicalHeaders strings.Builder
	for _, name := range headerNames {
		value := req.Header.Get(name)
		if name == "host" {
			value = req.URL.Host
		}
		canonicalHeaders.WriteString(name + ":" + strings.TrimSpace(value) + "\n")
	}
	signedHeaders := strings.Join(headerNames, ";")
	canonicalRequest := strings.Join([]string{req.Method, req.URL.EscapedPath(), canonicalQuery(req.URL.Query()), canonicalHeaders.String(), signedHeaders, payloadHash}, "\n")
	date := now.Format("20060102")
	scope := date + "/" + c.credentials.Region + "/s3/aws4_request"
	stringToSign := "AWS4-HMAC-SHA256\n" + now.Format("20060102T150405Z") + "\n" + scope + "\n" + sha256Hex([]byte(canonicalRequest))
	dateKey := hmacSHA256([]byte("AWS4"+c.credentials.SecretKey), date)
	regionKey := hmacSHA256(dateKey, c.credentials.Region)
	serviceKey := hmacSHA256(regionKey, "s3")
	signingKey := hmacSHA256(serviceKey, "aws4_request")
	signature := hex.EncodeToString(hmacSHA256(signingKey, stringToSign))
	req.Header.Set("Authorization", "AWS4-HMAC-SHA256 Credential="+c.credentials.AccessKey+"/"+scope+", SignedHeaders="+signedHeaders+", Signature="+signature)
}

func canonicalQuery(values url.Values) string {
	if len(values) == 0 {
		return ""
	}
	return values.Encode()
}
func sha256Hex(value []byte) string {
	sum := sha256.Sum256(value)
	return hex.EncodeToString(sum[:])
}
func hmacSHA256(key []byte, value string) []byte {
	mac := hmac.New(sha256.New, key)
	_, _ = mac.Write([]byte(value))
	return mac.Sum(nil)
}
