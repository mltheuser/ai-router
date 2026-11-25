package gitlab

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mltheuser/ai-router/httpx"
	"github.com/mltheuser/ai-router/router"
	"github.com/mltheuser/ai-router/router/chat"
)

// model is a model the gateway proxies, with the metadata Anthropic's Models
// API would report for it.
type model struct {
	id            string
	contextWindow int
	// maxTokens is the model's output limit, sent as max_tokens when a request
	// sets none.
	maxTokens int
}

// models are the models this provider serves. The proxy has no models
// endpoint, so they are listed by hand: each must be on the gateway's
// allow-list (the proxy_provider: anthropic entries in ai-assist's
// ai_gateway/model_selection/models.yml). The limits are Anthropic's, as its
// Models API reports them; the max_tokens in models.yml is GitLab's default
// for its own calls and can be lower.
var models = []model{
	{id: "claude-opus-5-5", contextWindow: 1_000_000, maxTokens: 128_000},
	{id: "claude-fable-5-2", contextWindow: 1_000_000, maxTokens: 128_000},
	{id: "claude-sonnet-5-5", contextWindow: 1_000_000, maxTokens: 128_000},
}

// features are the chat features of every listed model.
var features = []chat.Feature{
	chat.FeatureTools,
	chat.FeatureVision,
	chat.FeatureReasoning,
	chat.FeatureStructuredOutput,
}

// ListChatModels returns the hard-coded models. Pricing stays nil: GitLab
// bills in credits, not per token.
func (p *Provider) ListChatModels(_ context.Context) ([]chat.Model, error) {
	out := make([]chat.Model, 0, len(models))
	for _, m := range models {
		out = append(out, chat.Model{
			ModelRef:      router.NewModelRef(p, m.id),
			Features:      features,
			ContextWindow: m.contextWindow,
		})
	}
	return out, nil
}

// lookupModel returns the listed model with the given id.
func lookupModel(id string) (model, error) {
	for _, m := range models {
		if m.id == id {
			return m, nil
		}
	}
	return model{}, httpx.NewError(http.StatusNotFound, fmt.Sprintf("model '%s' is not served by gitlab", id))
}
