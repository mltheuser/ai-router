package openai

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/mltheuser/ai-router/api"
	"gopkg.in/yaml.v3"
)

// modelsResponse is the response from GET /v1/models. Each entry carries only
// an ID and ownership data — no capabilities, context window or pricing — so
// only the ID is decoded.
type modelsResponse struct {
	Data []openAIModel `json:"data"`
}

type openAIModel struct {
	ID string `json:"id"`
}

// modelMeta is the hand-maintained metadata for one model in models.yaml.
type modelMeta struct {
	Capabilities   []api.Capability `yaml:"capabilities"`
	ContextWindow  int              `yaml:"context_window"`
	CostPerMInput  *float64         `yaml:"cost_per_m_input"`
	CostPerMOutput *float64         `yaml:"cost_per_m_output"`
}

//go:embed models.yaml
var modelsYAML []byte

// knownModels maps model IDs to their metadata. The file is embedded at build
// time, so a parse failure is a programming error and panics at startup.
var knownModels = func() map[string]modelMeta {
	var models map[string]modelMeta
	if err := yaml.Unmarshal(modelsYAML, &models); err != nil {
		panic(fmt.Sprintf("parsing openai models.yaml: %v", err))
	}
	return models
}()

// ListModels fetches all models from OpenAI and enriches them with the
// metadata from models.yaml.
func (p *Provider) ListModels(ctx context.Context) ([]api.ModelInfo, error) {
	var resp modelsResponse
	if err := p.client.get(ctx, "/models", &resp); err != nil {
		return nil, fmt.Errorf("listing openai models: %w", err)
	}

	models := make([]api.ModelInfo, 0, len(resp.Data))
	for _, m := range resp.Data {
		models = append(models, convertModel(m, p.Name()))
	}
	return models, nil
}

// convertModel converts an OpenAI model to a unified ModelInfo. Models absent
// from models.yaml get an empty (non-nil, so it encodes as []) capability list
// and unknown context window and pricing: they are listed but not routable.
func convertModel(m openAIModel, providerName string) api.ModelInfo {
	info := api.ModelInfo{
		ID:           m.ID,
		Provider:     providerName,
		ProviderType: api.ProviderTypeCloud,
		Capabilities: []api.Capability{},
	}

	if meta, ok := knownModels[m.ID]; ok {
		info.Capabilities = meta.Capabilities
		info.ContextWindow = meta.ContextWindow
		info.CostPerMInput = meta.CostPerMInput
		info.CostPerMOutput = meta.CostPerMOutput
	}

	return info
}
