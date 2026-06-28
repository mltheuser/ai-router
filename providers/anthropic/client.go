package anthropic

import (
	"context"

	"github.com/mltheuser/ai-router/providers/httpclient"
)

const baseURL = "https://api.anthropic.com/v1"

// anthropicVersion pins the API version contract via the required
// anthropic-version header. It is a version identifier, not a release date:
// 2023-06-01 is the current stable value and rarely changes.
const anthropicVersion = "2023-06-01"

// client wraps HTTP requests to the Anthropic API using the shared httpclient.
type client struct {
	*httpclient.Client
}

func newClient(apiKey string) *client {
	return &client{
		Client: httpclient.New(
			baseURL,
			httpclient.WithHeader("x-api-key", apiKey),
			httpclient.WithHeader("anthropic-version", anthropicVersion),
		),
	}
}

// get is a convenience wrapper around the shared client's Get method.
func (c *client) get(ctx context.Context, path string, result interface{}) error {
	return c.Get(ctx, path, result)
}

// post is a convenience wrapper around the shared client's Post method.
func (c *client) post(ctx context.Context, path string, body interface{}, result interface{}) error {
	return c.Post(ctx, path, body, result)
}
