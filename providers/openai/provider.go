// Package openai is the provider for the OpenAI cloud API.
package openai

import (
	"context"
	"fmt"

	"github.com/mltheuser/ai-router/httpx"
	"github.com/mltheuser/ai-router/router"
	"github.com/mltheuser/ai-router/router/chat"
	"github.com/mltheuser/ai-router/router/embedding"
)

// Provider talks to the OpenAI API.
type Provider struct {
	client *httpx.Client
}

// The use cases this provider serves.
var (
	_ chat.Provider      = (*Provider)(nil)
	_ embedding.Provider = (*Provider)(nil)
)

// New creates a new OpenAI provider with the given API key.
func New(apiKey string) *Provider {
	return &Provider{
		client: newClient(apiKey),
	}
}

func (p *Provider) Name() string {
	return "openai"
}

func (p *Provider) Type() router.ProviderType {
	return router.Cloud
}

// Verify checks reachability and authentication. OpenAI has no dedicated
// key-info endpoint, so we list models — which requires a valid API key.
func (p *Provider) Verify(ctx context.Context) error {
	var resp modelsResponse
	if err := p.client.Get(ctx, "/models", &resp); err != nil {
		return fmt.Errorf("openai verification failed: %w", err)
	}
	return nil
}
