// Package rixlclient provides a minimal HTTP client for the Rixl API.
package rixlclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.rixl.com"

// Client is a thin HTTP client for the Rixl REST API.
type Client struct {
	baseURL    string
	httpClient *http.Client
	headers    http.Header
}

// Option configures a Client.
type Option func(*Client)

// WithHTTPClient overrides the default HTTP client.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(c *Client) {
		c.httpClient = httpClient
	}
}

// New creates a Rixl API client. At least one of apiKey or bearerToken must be
// provided; they correspond to the X-API-Key and Authorization: Bearer headers.
func New(apiKey, bearerToken, baseURL string, opts ...Option) (*Client, error) {
	if baseURL == "" {
		baseURL = defaultBaseURL
	}
	if !strings.HasPrefix(baseURL, "http") {
		return nil, fmt.Errorf("base_url must be an absolute URL, got %q", baseURL)
	}

	c := &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{Timeout: 60 * time.Second},
		headers:    make(http.Header),
	}
	for _, opt := range opts {
		opt(c)
	}

	if apiKey != "" {
		c.headers.Set("X-Api-Key", apiKey)
	}
	if bearerToken != "" {
		c.headers.Set("Authorization", "Bearer "+bearerToken)
	}

	return c, nil
}

// Get performs a GET request.
func (c *Client) Get(ctx context.Context, path string, query url.Values) ([]byte, error) {
	return c.Do(ctx, http.MethodGet, path, query, nil)
}

// Post performs a POST request.
func (c *Client) Post(ctx context.Context, path string, body []byte) ([]byte, error) {
	return c.Do(ctx, http.MethodPost, path, nil, body)
}

// Put performs a PUT request.
func (c *Client) Put(ctx context.Context, path string, body []byte) ([]byte, error) {
	return c.Do(ctx, http.MethodPut, path, nil, body)
}

// Patch performs a PATCH request.
func (c *Client) Patch(ctx context.Context, path string, body []byte) ([]byte, error) {
	return c.Do(ctx, http.MethodPatch, path, nil, body)
}

// Delete performs a DELETE request.
func (c *Client) Delete(ctx context.Context, path string) ([]byte, error) {
	return c.Do(ctx, http.MethodDelete, path, nil, nil)
}

// Do performs an HTTP request and returns the response body.
func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body []byte) ([]byte, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	u := c.baseURL + path
	if query != nil {
		u += "?" + query.Encode()
	}

	var bodyReader io.Reader
	if body != nil {
		bodyReader = bytes.NewReader(body)
	}

	req, err := http.NewRequestWithContext(ctx, method, u, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	for k, vals := range c.headers {
		for _, v := range vals {
			req.Header.Add(k, v)
		}
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return respBody, nil
	}

	if resp.StatusCode == http.StatusNotFound {
		return respBody, &NotFoundError{Path: path, Status: resp.StatusCode}
	}

	return respBody, fmt.Errorf("API error: %s %s: %d %s", method, path, resp.StatusCode, string(respBody))
}

// NotFoundError is returned when the API responds with HTTP 404.
type NotFoundError struct {
	Path   string
	Status int
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("not found: %s (HTTP %d)", e.Path, e.Status)
}

// IsNotFound reports whether err is a NotFoundError.
func IsNotFound(err error) bool {
	var e *NotFoundError
	return errors.As(err, &e)
}
