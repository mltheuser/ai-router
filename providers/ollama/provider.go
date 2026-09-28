// Package ollama is the provider for a local Ollama runner.
package ollama

import (
	"context"
	"net/http"

	"github.com/mltheuser/ai-router/provider"
	"github.com/mltheuser/ai-router/usecase/chat"
	"github.com/mltheuser/ai-router/usecase/embedding"
)

// Provider serves chat and embedding through a local Ollama runner.
type Provider struct {
	client *client
}

// The use cases this provider serves. A use case finds its providers by
// interface, so these checks turn a signature mismatch into a build error
// instead of a silently missing use case.
var (
	_ chat.Provider      = (*Provider)(nil)
	_ embedding.Provider = (*Provider)(nil)
)

// New creates a new Ollama provider pointing at the default local endpoint.
func New() *Provider {
	return &Provider{
		client: newClient("http://localhost:11434"),
	}
}

func (p *Provider) Name() string {
	return "ollama"
}

func (p *Provider) Type() provider.Type {
	return provider.Local
}

// Verify checks that Ollama is running and responding.
func (p *Provider) Verify(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.client.BaseURL+"/", nil)
	if err != nil {
		return err
	}

	resp, err := p.client.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	_ = resp.Body.Close()
	return nil
}
