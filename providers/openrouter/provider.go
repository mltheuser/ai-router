// Package openrouter is the provider for the OpenRouter cloud aggregator.
package openrouter

import (
	"context"
	"fmt"

	"github.com/mltheuser/ai-router/httpx"
	"github.com/mltheuser/ai-router/router"
	"github.com/mltheuser/ai-router/router/chat"
	"github.com/mltheuser/ai-router/router/embedding"
)

// Provider talks to the OpenRouter API.
type Provider struct {
	client *httpx.Client
}

// The use cases this provider serves.
var (
	_ chat.Provider      = (*Provider)(nil)
	_ embedding.Provider = (*Provider)(nil)
)

// New creates a new OpenRouter provider with the given API key.
func New(apiKey string) *Provider {
	return &Provider{
		client: newClient(apiKey),
	}
}

func (p *Provider) Name() string {
	return "openrouter"
}

func (p *Provider) Type() router.ProviderType {
	return router.Cloud
}

// keyResponse is the response from GET /api/v1/key.
type keyResponse struct {
	Data keyData `json:"data"`
}

type keyData struct {
	Label          string  `json:"label"`
	Usage          float64 `json:"usage"`
	Limit          float64 `json:"limit"`
	LimitRemaining float64 `json:"limit_remaining"`
	IsFreeTier     bool    `json:"is_free_tier"`
}

// Verify validates the API key by calling the dedicated key info endpoint.
func (p *Provider) Verify(ctx context.Context) error {
	var resp keyResponse
	if err := p.client.Get(ctx, "/key", &resp); err != nil {
		return fmt.Errorf("openrouter verification failed: %w", err)
	}
	return nil
}
