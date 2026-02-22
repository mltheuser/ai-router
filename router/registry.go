package router

import (
	"github.com/mltheuser/ai-router/provider"
	"github.com/mltheuser/ai-router/providers/ollama"
	"github.com/mltheuser/ai-router/providers/openrouter"
)

// DefaultRegistry returns the default configured registry with all known providers.
func DefaultRegistry() *provider.Registry {
	r := provider.NewRegistry()

	// Cloud providers
	r.RegisterCloud("openrouter", func(apiKey string) provider.Provider {
		return openrouter.New(apiKey)
	})

	// Local runners
	r.RegisterLocal("ollama", func() provider.Provider {
		return ollama.New()
	})

	return r
}
