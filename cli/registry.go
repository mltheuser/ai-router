package cli

import (
	"context"

	"github.com/mltheuser/ai-router/provider"
	"github.com/mltheuser/ai-router/providers/anthropic"
	"github.com/mltheuser/ai-router/providers/ollama"
	"github.com/mltheuser/ai-router/providers/openai"
	"github.com/mltheuser/ai-router/providers/openrouter"
	"github.com/mltheuser/ai-router/usecase"
	"github.com/mltheuser/ai-router/usecase/chat"
	"github.com/mltheuser/ai-router/usecase/embedding"
)

// This file is the registry: the one place that knows every provider and
// every use case. Providers and use cases know nothing of each other beyond
// the use-case interfaces; they meet here.

// cloudProviders builds each cloud provider from its API key, which is read
// from AI_ROUTER_<NAME>_API_KEY. A provider whose key is unset is skipped.
var cloudProviders = map[string]func(apiKey string) provider.Provider{
	"anthropic":  func(key string) provider.Provider { return anthropic.New(key) },
	"openai":     func(key string) provider.Provider { return openai.New(key) },
	"openrouter": func(key string) provider.Provider { return openrouter.New(key) },
}

// localProviders builds each local provider. One that is not running is
// skipped.
var localProviders = []func() provider.Provider{
	func() provider.Provider { return ollama.New() },
}

// useCases builds every use case over the verified providers. Each serves the
// providers that implement its interface and lists their models once.
func useCases(ctx context.Context, providers []provider.Provider) []usecase.UseCase {
	return []usecase.UseCase{
		chat.New(ctx, providers),
		embedding.New(ctx, providers),
	}
}
