// Package provider defines the Provider interface that every backend
// implements, plus the registry machinery that maps provider names to
// factories.
package provider

import (
	"context"

	"github.com/mltheuser/ai-router/api"
)

// Provider is the contract that every cloud and local provider must implement.
// Methods should return api.ErrNotSupported for unsupported capabilities.
type Provider interface {
	// Name returns the provider identifier (e.g. "openrouter", "ollama").
	Name() string

	// Type returns whether this is a cloud or local provider.
	Type() api.ProviderType

	// Verify checks that the provider is reachable and properly authenticated.
	// For cloud providers this validates the API key; for local providers this
	// checks that the runner is running and responding.
	Verify(ctx context.Context) error

	// ListModels returns all models available through this provider.
	ListModels(ctx context.Context) ([]api.ModelInfo, error)

	// Embed generates embeddings for the given input.
	Embed(ctx context.Context, req *api.EmbedRequest) (*api.EmbedResponse, error)

	// Chat generates a chat completion for the given request.
	Chat(ctx context.Context, req *api.ChatRequest) (*api.ChatResponse, error)
}
