// Package openai implements the Provider interface for the OpenAI cloud API.
package openai

import (
	"context"
	"fmt"

	"github.com/mltheuser/ai-router/api"
)

// Provider implements the provider.Provider interface for OpenAI.
type Provider struct {
	client *client
}

// New creates a new OpenAI provider with the given API key.
func New(apiKey string) *Provider {
	return &Provider{
		client: newClient(apiKey),
	}
}

func (p *Provider) Name() string {
	return "openai"
}

func (p *Provider) Type() api.ProviderType {
	return api.ProviderTypeCloud
}

// Verify checks reachability and authentication. OpenAI has no dedicated
// key-info endpoint, so we list models — which requires a valid API key.
func (p *Provider) Verify(ctx context.Context) error {
	var resp modelsResponse
	if err := p.client.get(ctx, "/models", &resp); err != nil {
		return fmt.Errorf("openai verification failed: %w", err)
	}
	return nil
}
