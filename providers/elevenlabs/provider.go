// Package elevenlabs is the provider for the ElevenLabs API.
//
// It expects a pay-as-you-go API key with the permissions models_read,
// voices_read, text_to_speech and speech_to_text. The voices it serves are
// those in the account's My Voices; add a Voice Library voice there to serve
// it.
package elevenlabs

import (
	"context"
	"fmt"

	"github.com/mltheuser/ai-router/httpx"
	"github.com/mltheuser/ai-router/router"
	"github.com/mltheuser/ai-router/router/speechtotext"
	"github.com/mltheuser/ai-router/router/texttospeech"
)

// Provider talks to the ElevenLabs API.
type Provider struct {
	client *httpx.Client
}

// The use cases this provider serves.
var (
	_ texttospeech.Provider = (*Provider)(nil)
	_ speechtotext.Provider = (*Provider)(nil)
)

// New creates a new ElevenLabs provider with the given API key.
func New(apiKey string) *Provider {
	return &Provider{
		client: newClient(apiKey),
	}
}

func (p *Provider) Name() string {
	return "elevenlabs"
}

func (p *Provider) Type() router.ProviderType {
	return router.Cloud
}

// Verify checks reachability and authentication by listing models, which
// requires the API key.
func (p *Provider) Verify(ctx context.Context) error {
	if err := p.client.Get(ctx, "/v1/models", nil); err != nil {
		return fmt.Errorf("elevenlabs verification failed: %w", err)
	}
	return nil
}
