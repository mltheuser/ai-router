// Package ollama is the provider for a local Ollama runner.
package ollama

import (
	"context"

	"github.com/mltheuser/ai-router/httpx"
	"github.com/mltheuser/ai-router/router"
	"github.com/mltheuser/ai-router/router/chat"
	"github.com/mltheuser/ai-router/router/embedding"
)

// Provider talks to a local Ollama runner.
type Provider struct {
	client *httpx.Client
}

// The use cases this provider serves.
var (
	_ chat.Provider      = (*Provider)(nil)
	_ embedding.Provider = (*Provider)(nil)
)

// New creates a new Ollama provider pointing at the default local endpoint.
func New() *Provider {
	return &Provider{
		client: httpx.NewClient("http://localhost:11434"),
	}
}

func (p *Provider) Name() string {
	return "ollama"
}

func (p *Provider) Type() router.ProviderType {
	return router.Local
}

func (p *Provider) Verify(ctx context.Context) error {
	return p.client.Get(ctx, "/api/version", nil)
}
