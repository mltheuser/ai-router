package anthropic

import (
	"context"

	"github.com/mltheuser/ai-router/api"
)

// Embed is unsupported: Anthropic offers no embeddings API.
func (p *Provider) Embed(_ context.Context, _ *api.EmbedRequest) (*api.EmbedResponse, error) {
	return nil, api.ErrNotSupported
}
