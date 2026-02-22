package ollama

import (
	"context"
	"net/http"

	"github.com/mltheuser/ai-router/api"
)

// Provider implements provider.Provider for Ollama.
type Provider struct {
	client *client
}

// New creates a new Ollama provider pointing at the default local endpoint.
func New() *Provider {
	return &Provider{
		client: newClient("http://localhost:11434"),
	}
}

func (p *Provider) Name() string {
	return "ollama"
}

func (p *Provider) Type() api.ProviderType {
	return api.ProviderTypeLocal
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
	resp.Body.Close()
	return nil
}
