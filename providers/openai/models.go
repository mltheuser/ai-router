package openai

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/mltheuser/ai-router/provider"
	"github.com/mltheuser/ai-router/usecase/chat"
	"github.com/mltheuser/ai-router/usecase/embedding"
	"gopkg.in/yaml.v3"
)

// modelsResponse is the response from GET /v1/models. Each entry carries only
// an ID and ownership data — no capabilities, context window or pricing — so
// only the ID is decoded.
type modelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

// chatMeta and embeddingMeta are the hand-maintained metadata of one model in
// models.yaml.
type chatMeta struct {
	Features       []chat.Feature `yaml:"features"`
	ContextWindow  int            `yaml:"context_window"`
	CostPerMInput  *float64       `yaml:"cost_per_m_input"`
	CostPerMOutput *float64       `yaml:"cost_per_m_output"`
}

type embeddingMeta struct {
	ContextWindow int      `yaml:"context_window"`
	CostPerMInput *float64 `yaml:"cost_per_m_input"`
}

//go:embed models.yaml
var modelsYAML []byte

// knownModels is models.yaml, by use case and then by model ID. The file is
// embedded at build time, so a parse failure is a programming error and
// panics at startup.
var knownModels = func() (known struct {
	Chat      map[string]chatMeta      `yaml:"chat"`
	Embedding map[string]embeddingMeta `yaml:"embedding"`
},
) {
	if err := yaml.Unmarshal(modelsYAML, &known); err != nil {
		panic(fmt.Sprintf("parsing openai models.yaml: %v", err))
	}
	return known
}()

// ListChatModels returns the available models that models.yaml describes as
// chat models.
func (p *Provider) ListChatModels(ctx context.Context) ([]chat.Model, error) {
	ids, err := p.listModelIDs(ctx)
	if err != nil {
		return nil, err
	}
	var models []chat.Model
	for _, id := range ids {
		if meta, ok := knownModels.Chat[id]; ok {
			models = append(models, chat.Model{
				ModelRef:       provider.NewModelRef(p, id),
				Features:       meta.Features,
				ContextWindow:  meta.ContextWindow,
				CostPerMInput:  meta.CostPerMInput,
				CostPerMOutput: meta.CostPerMOutput,
			})
		}
	}
	return models, nil
}

// ListEmbeddingModels returns the available models that models.yaml
// describes as embedding models.
func (p *Provider) ListEmbeddingModels(ctx context.Context) ([]embedding.Model, error) {
	ids, err := p.listModelIDs(ctx)
	if err != nil {
		return nil, err
	}
	var models []embedding.Model
	for _, id := range ids {
		if meta, ok := knownModels.Embedding[id]; ok {
			models = append(models, embedding.Model{
				ModelRef:      provider.NewModelRef(p, id),
				ContextWindow: meta.ContextWindow,
				CostPerMInput: meta.CostPerMInput,
			})
		}
	}
	return models, nil
}

// listModelIDs returns the IDs of all models the API key can use.
func (p *Provider) listModelIDs(ctx context.Context) ([]string, error) {
	var resp modelsResponse
	if err := p.client.Get(ctx, "/models", &resp); err != nil {
		return nil, fmt.Errorf("listing openai models: %w", err)
	}
	ids := make([]string, len(resp.Data))
	for i, m := range resp.Data {
		ids[i] = m.ID
	}
	return ids, nil
}
