package ollama

import (
	"context"

	"github.com/mltheuser/ai-router/providers/httpclient"
)

// client wraps HTTP requests to the Ollama API using the shared httpclient.
type client struct {
	*httpclient.Client
}

func newClient(endpoint string) *client {
	return &client{
		Client: httpclient.New(endpoint),
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
