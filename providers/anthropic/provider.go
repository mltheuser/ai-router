// Package anthropic is the provider for the Anthropic (Claude) cloud API.
package anthropic

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mltheuser/ai-router/httpx"
	"github.com/mltheuser/ai-router/router"
	"github.com/mltheuser/ai-router/router/chat"
)

// Provider talks to the Anthropic API.
type Provider struct {
	client *httpx.Client
}

// The use cases this provider serves.
var (
	_ chat.Provider = (*Provider)(nil)
)

// New creates a new Anthropic provider with the given API key.
func New(apiKey string) *Provider {
	return &Provider{
		client: newClient(apiKey),
	}
}

// modelMaxTokens returns the model's maximum output-token count.
func (p *Provider) modelMaxTokens(ctx context.Context, model string) int {
	var m anthropicModel
	if err := p.client.Get(ctx, "/models/"+url.PathEscape(model), &m); err != nil || m.MaxTokens <= 0 {
		return fallbackMaxTokens
	}
	return m.MaxTokens
}

func (p *Provider) Name() string {
	return "anthropic"
}

func (p *Provider) Type() router.ProviderType {
	return router.Cloud
}

// Verify checks reachability and authentication. Anthropic has no dedicated
// key-info endpoint, so we issue a minimal models listing request — which does
// require the API key.
func (p *Provider) Verify(ctx context.Context) error {
	var resp modelsResponse
	if err := p.client.Get(ctx, "/models?limit=1", &resp); err != nil {
		return fmt.Errorf("anthropic verification failed: %w", err)
	}
	return nil
}
