// Package exa is the provider for the Exa web search API.
package exa

import (
	"context"
	"fmt"

	"github.com/mltheuser/ai-router/httpx"
	"github.com/mltheuser/ai-router/router"
	"github.com/mltheuser/ai-router/router/contents"
	"github.com/mltheuser/ai-router/router/search"
)

// Provider serves search and contents through the Exa API.
type Provider struct {
	client *httpx.Client
}

// The use cases this provider serves. A use case finds its providers by
// interface, so these checks turn a signature mismatch into a build error
// instead of a silently missing use case.
var (
	_ search.Provider   = (*Provider)(nil)
	_ contents.Provider = (*Provider)(nil)
)

// New creates a new Exa provider with the given API key.
func New(apiKey string) *Provider {
	return &Provider{
		client: newClient(apiKey),
	}
}

func (p *Provider) Name() string {
	return "exa"
}

func (p *Provider) Type() router.ProviderType {
	return router.Cloud
}

// Verify checks reachability and authentication. Exa has no key-info
// endpoint, so we list monitors: it requires the API key and is free.
func (p *Provider) Verify(ctx context.Context) error {
	if err := p.client.Get(ctx, "/monitors", nil); err != nil {
		return fmt.Errorf("exa verification failed: %w", err)
	}
	return nil
}
