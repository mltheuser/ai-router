// Package openrouter implements the Provider interface for the OpenRouter
// cloud aggregator.
package openrouter

import (
	"context"
	"fmt"
	"sync"

	"github.com/mltheuser/ai-router/api"
)

// Provider implements the provider.Provider interface for OpenRouter.
type Provider struct {
	client *client

	// Internal metadata to track if a model supports "reasoning_effort".
	// This is populated during ListModels.
	mu                      sync.RWMutex
	supportsReasoningEffort map[string]bool
}

// New creates a new OpenRouter provider with the given API key.
func New(apiKey string) *Provider {
	return &Provider{
		client:                  newClient(apiKey),
		supportsReasoningEffort: make(map[string]bool),
	}
}

func (p *Provider) Name() string {
	return "openrouter"
}

func (p *Provider) Type() api.ProviderType {
	return api.ProviderTypeCloud
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
	if err := p.client.get(ctx, "/key", &resp); err != nil {
		return fmt.Errorf("invalid API key: %w", err)
	}
	return nil
}
