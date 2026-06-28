package anthropic

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mltheuser/ai-router/api"
)

// modelsResponse is the response from GET /v1/models.
type modelsResponse struct {
	Data    []anthropicModel `json:"data"`
	FirstID string           `json:"first_id"`
	HasMore bool             `json:"has_more"`
	LastID  string           `json:"last_id"`
}

type anthropicModel struct {
	ID             string                `json:"id"`
	Type           string                `json:"type"`
	DisplayName    string                `json:"display_name"`
	CreatedAt      string                `json:"created_at"`
	MaxInputTokens int                   `json:"max_input_tokens"`
	MaxTokens      int                   `json:"max_tokens"`
	Capabilities   anthropicCapabilities `json:"capabilities"`
}

// anthropicCapabilities models only the capability sub-fields we use.
type anthropicCapabilities struct {
	ImageInput        anthropicCapability `json:"image_input"`
	Thinking          anthropicCapability `json:"thinking"`
	StructuredOutputs anthropicCapability `json:"structured_outputs"`
}

type anthropicCapability struct {
	Supported bool `json:"supported"`
}

// ListModels fetches all models from Anthropic and converts them to unified
// ModelInfo. The endpoint is paginated via opaque cursors, so we follow
// has_more/last_id until the listing is exhausted.
func (p *Provider) ListModels(ctx context.Context) ([]api.ModelInfo, error) {
	var models []api.ModelInfo

	afterID := ""
	for {
		query := url.Values{}
		query.Set("limit", "1000")
		if afterID != "" {
			query.Set("after_id", afterID)
		}

		var resp modelsResponse
		if err := p.client.get(ctx, "/models?"+query.Encode(), &resp); err != nil {
			return nil, fmt.Errorf("listing anthropic models: %w", err)
		}

		for _, m := range resp.Data {
			models = append(models, convertModel(m, p.Name()))
		}

		// Terminate on the last page, an empty page, or a missing cursor to
		// guarantee the loop always ends.
		if !resp.HasMore || len(resp.Data) == 0 || resp.LastID == "" {
			break
		}
		afterID = resp.LastID
	}

	return models, nil
}

// convertModel converts an Anthropic model to a unified ModelInfo. The models
// endpoint exposes no pricing, so the cost fields are left nil.
func convertModel(m anthropicModel, providerName string) api.ModelInfo {
	return api.ModelInfo{
		ID:            m.ID,
		Provider:      providerName,
		ProviderType:  api.ProviderTypeCloud,
		Capabilities:  inferCapabilities(m),
		ContextWindow: m.MaxInputTokens,
	}
}

// inferCapabilities maps Anthropic capability flags to unified capabilities.
// The models API exposes no chat or tools flag, so we assume both — every
// Claude model supports them — and read the rest from the reported flags.
func inferCapabilities(m anthropicModel) []api.Capability {
	caps := []api.Capability{api.CapabilityChat, api.CapabilityTools}

	if m.Capabilities.ImageInput.Supported {
		caps = append(caps, api.CapabilityVision)
	}
	if m.Capabilities.Thinking.Supported {
		caps = append(caps, api.CapabilityReasoning)
	}
	if m.Capabilities.StructuredOutputs.Supported {
		caps = append(caps, api.CapabilityStructuredOutput)
	}

	return caps
}
