package cli

import (
	"context"

	"github.com/mltheuser/ai-router/providers/anthropic"
	"github.com/mltheuser/ai-router/providers/exa"
	"github.com/mltheuser/ai-router/providers/ollama"
	"github.com/mltheuser/ai-router/providers/openai"
	"github.com/mltheuser/ai-router/providers/openrouter"
	"github.com/mltheuser/ai-router/router"
	"github.com/mltheuser/ai-router/router/chat"
	"github.com/mltheuser/ai-router/router/contents"
	"github.com/mltheuser/ai-router/router/embedding"
	"github.com/mltheuser/ai-router/router/search"
)

// cloudProviders builds each cloud provider from its API key.
var cloudProviders = map[string]func(apiKey string) router.Provider{
	"anthropic":  func(key string) router.Provider { return anthropic.New(key) },
	"exa":        func(key string) router.Provider { return exa.New(key) },
	"openai":     func(key string) router.Provider { return openai.New(key) },
	"openrouter": func(key string) router.Provider { return openrouter.New(key) },
}

// localProviders builds each local provider.
var localProviders = []func() router.Provider{
	func() router.Provider { return ollama.New() },
}

// useCases builds every use case over the verified providers.
func useCases(ctx context.Context, providers []router.Provider) []router.UseCase {
	return []router.UseCase{
		chat.New(ctx, providers),
		embedding.New(ctx, providers),
		search.New(ctx, providers),
		contents.New(ctx, providers),
	}
}
