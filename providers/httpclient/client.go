package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// --- Debug collector (context-based) ---

type debugCtxKey struct{}

// DebugCollector captures raw HTTP request/response bodies for debug logging.
// It is passed through context so providers don't need any code changes.
type DebugCollector struct {
	RequestMethod string
	RequestURL    string
	RequestBody   []byte
	ResponseBody  []byte
}

// NewDebugContext returns a child context carrying the given collector.
func NewDebugContext(ctx context.Context, dc *DebugCollector) context.Context {
	return context.WithValue(ctx, debugCtxKey{}, dc)
}

// DebugCollectorFromContext extracts the collector, or nil if absent.
func DebugCollectorFromContext(ctx context.Context) *DebugCollector {
	dc, _ := ctx.Value(debugCtxKey{}).(*DebugCollector)
	return dc
}

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

	// Capture request body for debug if collector is present
	if dc := DebugCollectorFromContext(ctx); dc != nil {
		dc.RequestBody = data
	}

	return c.Do(req, result)
}

// Do executes the given HTTP request, handles the response, and unmarshals the body into result (if not nil).
func (c *Client) Do(req *http.Request, result interface{}) error {
	for k, v := range c.Headers {
		req.Header.Set(k, v)
	}

	// Capture request metadata for debug if collector is present
	if dc := DebugCollectorFromContext(req.Context()); dc != nil {
		dc.RequestMethod = req.Method
		dc.RequestURL = req.URL.String()
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	// Capture response body for debug if collector is present
	if dc := DebugCollectorFromContext(req.Context()); dc != nil {
		dc.ResponseBody = body
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	if result != nil {
		if err := json.Unmarshal(body, result); err != nil {
			return fmt.Errorf("parsing response: %w", err)
		}
	}
	return nil
}
