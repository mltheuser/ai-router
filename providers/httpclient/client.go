// Package httpclient provides a small JSON-over-HTTP client shared by the
// provider implementations. It records every exchange into the request's
// debug.Exchange, if one is attached, so debug logging needs no provider code.
package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mltheuser/ai-router/api"
	"github.com/mltheuser/ai-router/debug"
)

// --- Client ---

// Client wraps HTTP requests with common logic for JSON APIs.
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	Headers    map[string]string
}

// Option allows configuring the Client.
type Option func(*Client)

// WithHeader adds a default header to all requests.
func WithHeader(key, value string) Option {
	return func(c *Client) {
		c.Headers[key] = value
	}
}

// New creates a new Client with the given base URL and options.
func New(baseURL string, options ...Option) *Client {
	c := &Client{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 10 * time.Minute,
		},
		Headers: make(map[string]string),
	}
	for _, opt := range options {
		opt(c)
	}
	return c
}

// Get sends a GET request to the specified path.
func (c *Client) Get(ctx context.Context, path string, result interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+path, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	return c.Do(req, result)
}

// Post sends a POST request with the given body to the specified path.
func (c *Client) Post(ctx context.Context, path string, body interface{}, result interface{}) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+path, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if ex := debug.ExchangeFrom(ctx); ex != nil {
		ex.RequestBody = data
	}

	return c.Do(req, result)
}

// Do executes the given HTTP request, handles the response, and unmarshals the body into result (if not nil).
func (c *Client) Do(req *http.Request, result interface{}) error {
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	ex := debug.ExchangeFrom(req.Context())
	if ex != nil {
		ex.Method = req.Method
		ex.URL = req.URL.String()
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if ex != nil {
		ex.ResponseBody = body
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return api.NewUpstreamError(resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}
	}
	return nil
}
