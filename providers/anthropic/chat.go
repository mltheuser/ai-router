package anthropic

import (
	"context"

	"github.com/mltheuser/ai-router/api"
)

// Chat is unsupported for now: this milestone wires only the models endpoint.
func (p *Provider) Chat(_ context.Context, _ *api.ChatRequest) (*api.ChatResponse, error) {
	return nil, api.ErrNotSupported
}
