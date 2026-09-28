package openrouter

import (
	"context"
	"fmt"
	"slices"
	"strconv"

	"github.com/mltheuser/ai-router/provider"
	"github.com/mltheuser/ai-router/usecase/chat"
	"github.com/mltheuser/ai-router/usecase/embedding"
)

// modelsResponse is the response of both GET /models (chat models) and
// GET /embeddings/models.
type modelsResponse struct {
	Data []openRouterModel `json:"data"`
}

type openRouterModel struct {
	ID                  string                 `json:"id"`
	ContextLength       int                    `json:"context_length"`
	Architecture        openRouterArchitecture `json:"architecture"`
	Pricing             openRouterPricing      `json:"pricing"`
	SupportedParameters []string               `json:"supported_parameters"`
}

type openRouterArchitecture struct {
	InputModalities  []string `json:"input_modalities"`
	OutputModalities []string `json:"output_modalities"`
}

type openRouterPricing struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
}

// ListChatModels returns the models that produce text.
func (p *Provider) ListChatModels(ctx context.Context) ([]chat.Model, error) {
	var resp modelsResponse
	if err := p.client.Get(ctx, "/models", &resp); err != nil {
		return nil, fmt.Errorf("listing openrouter chat models: %w", err)
	}

	var models []chat.Model
	for _, m := range resp.Data {
		if !slices.Contains(m.Architecture.OutputModalities, "text") {
			continue
		}
		models = append(models, chat.Model{
			ModelRef:       provider.NewModelRef(p, m.ID),
			Features:       features(m),
			ContextWindow:  m.ContextLength,
			CostPerMInput:  parsePrice(m.Pricing.Prompt),
			CostPerMOutput: parsePrice(m.Pricing.Completion),
		})
	}
	return models, nil
}

// ListEmbeddingModels returns the embedding models, which OpenRouter lists
// at an endpoint of their own.
func (p *Provider) ListEmbeddingModels(ctx context.Context) ([]embedding.Model, error) {
	var resp modelsResponse
	if err := p.client.Get(ctx, "/embeddings/models", &resp); err != nil {
		return nil, fmt.Errorf("listing openrouter embedding models: %w", err)
	}

	models := make([]embedding.Model, 0, len(resp.Data))
	for _, m := range resp.Data {
		models = append(models, embedding.Model{
			ModelRef:      provider.NewModelRef(p, m.ID),
			ContextWindow: m.ContextLength,
			CostPerMInput: parsePrice(m.Pricing.Prompt),
		})
	}
	return models, nil
}

// features derives a chat model's features from its input modalities and
// supported request parameters.
func features(m openRouterModel) []chat.Feature {
	var fs []chat.Feature
	if slices.Contains(m.Architecture.InputModalities, "image") {
		fs = append(fs, chat.FeatureVision)
	}
	for _, param := range m.SupportedParameters {
		switch param {
		case "structured_outputs":
			fs = append(fs, chat.FeatureStructuredOutput)
		case "reasoning":
			fs = append(fs, chat.FeatureReasoning)
		case "tools":
			fs = append(fs, chat.FeatureTools)
		}
	}
	return fs
}

// parsePrice converts a per-token price string to cost per million tokens.
// Router models (e.g. openrouter/auto) report "-1" because their price depends
// on the model they forward to; negative and unparsable prices are treated as
// unknown (nil).
func parsePrice(perToken string) *float64 {
	cost, err := strconv.ParseFloat(perToken, 64)
	if err != nil || cost < 0 {
		return nil
	}
	perMillion := cost * 1_000_000
	return &perMillion
}
