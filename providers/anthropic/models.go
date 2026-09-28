package anthropic

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mltheuser/ai-router/provider"
	"github.com/mltheuser/ai-router/usecase/chat"
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

// ListChatModels fetches all models from Anthropic. The endpoint is paginated
// via opaque cursors, so we follow has_more/last_id until the listing is
// exhausted.
func (p *Provider) ListChatModels(ctx context.Context) ([]chat.Model, error) {
	var models []chat.Model

	afterID := ""
	for {
		query := url.Values{}
		query.Set("limit", "1000")
		if afterID != "" {
			query.Set("after_id", afterID)
		}

		var resp modelsResponse
		if err := p.client.Get(ctx, "/models?"+query.Encode(), &resp); err != nil {
			return nil, fmt.Errorf("listing anthropic models: %w", err)
		}

		for _, m := range resp.Data {
			models = append(models, p.convertModel(m))
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

// convertModel converts an Anthropic model to a chat model. The models
// endpoint exposes no pricing, so the cost fields are left nil.
func (p *Provider) convertModel(m anthropicModel) chat.Model {
	return chat.Model{
		ModelRef:      provider.NewModelRef(p, m.ID),
		Features:      features(m),
		ContextWindow: m.MaxInputTokens,
	}
}

// features maps Anthropic capability flags to chat features. The models API
// exposes no tools flag, so we assume it: every Claude model supports tools.
func features(m anthropicModel) []chat.Feature {
	fs := []chat.Feature{chat.FeatureTools}

	if m.Capabilities.ImageInput.Supported {
		fs = append(fs, chat.FeatureVision)
	}
	if m.Capabilities.Thinking.Supported {
		fs = append(fs, chat.FeatureReasoning)
	}
	if m.Capabilities.StructuredOutputs.Supported {
		fs = append(fs, chat.FeatureStructuredOutput)
	}

	return fs
}
