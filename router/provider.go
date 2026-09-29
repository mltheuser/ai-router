package router

import (
	"context"
	"fmt"
)

// Provider is the contract every cloud and local provider implements.
type Provider interface {
	// Name returns the provider identifier (e.g. "openrouter", "ollama").
	Name() string

	// Type returns whether this is a cloud or local provider.
	Type() ProviderType

	// Verify checks that the provider is reachable and properly authenticated.
	// For cloud providers this validates the API key; for local providers this
	// checks that the runner is running and responding.
	Verify(ctx context.Context) error
}

// ProviderType distinguishes cloud from local providers.
type ProviderType string

// Provider types.
const (
	Cloud ProviderType = "cloud"
	Local ProviderType = "local"
)

// ModelRef is the routing identity of one model at one provider.
type ModelRef struct {
	ID string `json:"id"`
	// Model is the string to pass as "model" in requests.
	Model        string       `json:"model"`
	Provider     string       `json:"provider"`
	ProviderType ProviderType `json:"provider_type"`
}

// NewModelRef builds the ref of the model id served by p.
func NewModelRef(p Provider, id string) ModelRef {
	return ModelRef{
		ID:           id,
		Model:        fmt.Sprintf("%s:%s@%s", id, p.Type(), p.Name()),
		Provider:     p.Name(),
		ProviderType: p.Type(),
	}
}

// Ref returns the ref itself. Embedding ModelRef promotes this method, which
// is how generic code reads the ref of any use case's model type.
func (r ModelRef) Ref() ModelRef { return r }
