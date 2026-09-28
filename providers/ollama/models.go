package ollama

import (
	"context"
	"fmt"
	"slices"

	"github.com/mltheuser/ai-router/provider"
	"github.com/mltheuser/ai-router/usecase/chat"
	"github.com/mltheuser/ai-router/usecase/embedding"
)

// tagsResponse is the response from GET /api/tags.
type tagsResponse struct {
	Models []struct {
		Name string `json:"name"`
		Size int64  `json:"size"`
	} `json:"models"`
}

// showRequest and showResponse are the request and the part of the response
// we use of POST /api/show.
type showRequest struct {
	Name string `json:"name"`
}

type showResponse struct {
	Capabilities []string `json:"capabilities"`
}

// installedModel is an installed model with its Ollama capabilities
// ("completion", "embedding", "tools", "vision", "thinking", ...).
type installedModel struct {
	name         string
	size         int64
	capabilities []string
}

// ListChatModels returns the installed models that can complete text.
func (p *Provider) ListChatModels(ctx context.Context) ([]chat.Model, error) {
	installed, err := p.listInstalled(ctx)
	if err != nil {
		return nil, err
	}

	var models []chat.Model
	for _, m := range installed {
		free := 0.0
		if !slices.Contains(m.capabilities, "completion") {
			continue
		}
		// Ollama constrains any completion model's output to a JSON schema.
		features := []chat.Feature{chat.FeatureStructuredOutput}
		for _, c := range m.capabilities {
			switch c {
			case "tools":
				features = append(features, chat.FeatureTools)
			case "vision":
				features = append(features, chat.FeatureVision)
			case "thinking":
				features = append(features, chat.FeatureReasoning)
			}
		}
		models = append(models, chat.Model{
			ModelRef:       provider.NewModelRef(p, m.name),
			Features:       features,
			SizeBytes:      &m.size,
			CostPerMInput:  &free,
			CostPerMOutput: &free,
		})
	}
	return models, nil
}

// ListEmbeddingModels returns the installed models that can embed text.
func (p *Provider) ListEmbeddingModels(ctx context.Context) ([]embedding.Model, error) {
	installed, err := p.listInstalled(ctx)
	if err != nil {
		return nil, err
	}

	var models []embedding.Model
	for _, m := range installed {
		if slices.Contains(m.capabilities, "embedding") {
			free := 0.0
			models = append(models, embedding.Model{
				ModelRef:      provider.NewModelRef(p, m.name),
				SizeBytes:     &m.size,
				CostPerMInput: &free,
			})
		}
	}
	return models, nil
}

// listInstalled returns the installed models with their capabilities, which
// only /api/show reports, one model at a time.
func (p *Provider) listInstalled(ctx context.Context) ([]installedModel, error) {
	var tags tagsResponse
	if err := p.client.Get(ctx, "/api/tags", &tags); err != nil {
		return nil, fmt.Errorf("listing ollama models: %w", err)
	}

	models := make([]installedModel, 0, len(tags.Models))
	for _, m := range tags.Models {
		var show showResponse
		if err := p.client.Post(ctx, "/api/show", showRequest{Name: m.Name}, &show); err != nil {
			// Without its capabilities, assume a plain completion model.
			show.Capabilities = []string{"completion"}
		}
		models = append(models, installedModel{name: m.Name, size: m.Size, capabilities: show.Capabilities})
	}
	return models, nil
}
