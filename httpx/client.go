package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/mltheuser/ai-router/debug"
)

// Client makes JSON requests to one upstream API. It records every exchange
// into the request's debug.Exchange, if one is attached, so debug logging
// needs no code in the callers.
type Client struct {
	baseURL string
	headers map[string]string
	http    *http.Client
}

// Option configures a Client.
type Option func(*Client)

// WithHeader sends a header with every request, e.g. for authentication.
func WithHeader(key, value string) Option {
	return func(c *Client) { c.headers[key] = value }
}

// NewClient creates a Client for the API at baseURL. Paths passed to Get and
// Post are appended to it.
func NewClient(baseURL string, options ...Option) *Client {
	c := &Client{
		baseURL: baseURL,
		headers: make(map[string]string),
		http:    &http.Client{Timeout: 10 * time.Minute},
	}
	for _, opt := range options {
		opt(c)
	}
	return c
}

// Get sends a GET request and decodes the JSON response into result, unless
// result is nil.
func (c *Client) Get(ctx context.Context, path string, result any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+path, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	return c.do(req, result)
}

// Post sends body as JSON and decodes the JSON response into result, unless
// result is nil.
func (c *Client) Post(ctx context.Context, path string, body, result any) error {
	data, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+path, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	if ex := debug.ExchangeFrom(ctx); ex != nil {
		ex.RequestBody = data
	}
	return c.do(req, result)
}

// do sends req. A non-2xx response becomes an *Error via NewUpstreamError.
func (c *Client) do(req *http.Request, result any) error {
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	ex := debug.ExchangeFrom(req.Context())
	if ex != nil {
		ex.Method = req.Method
		ex.URL = req.URL.String()
	}

	resp, err := c.http.Do(req)
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
		return NewUpstreamError(resp.StatusCode, string(body))
	}
	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}
	}
	return nil
}
