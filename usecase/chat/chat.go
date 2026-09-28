// Package chat is the chat use case: a conversation of messages in, the
// model's next message out, optionally with tool calls, images, reasoning and
// structured output. It owns the whole contract: the model type listed at
// GET /v1/chat/models, the request and response of POST /v1/chat, the
// Provider interface a backend implements to serve them, and the scenarios
// that verify an implementation.
package chat

import (
	"context"
	"fmt"
	"net/http"
	"slices"

	"github.com/mltheuser/ai-router/api"
	"github.com/mltheuser/ai-router/provider"
	"github.com/mltheuser/ai-router/usecase"
)

// Provider is implemented by every backend that serves chat.
type Provider interface {
	provider.Provider

	// ListChatModels returns the chat models available at this provider.
	ListChatModels(ctx context.Context) ([]Model, error)

	// Chat generates the next message of the conversation. req.Model is the
	// bare model ID, as listed by ListChatModels.
	Chat(ctx context.Context, req *Request) (*Response, error)
}

// Feature is an optional ability of a chat model beyond plain text chat.
type Feature string

// Chat features.
const (
	FeatureTools            Feature = "tools"
	FeatureVision           Feature = "vision"
	FeatureReasoning        Feature = "reasoning"
	FeatureStructuredOutput Feature = "structured_output"
)

// Model describes a chat model at one provider.
type Model struct {
	provider.ModelRef
	Features []Feature `json:"features,omitempty"`

	// Cloud metadata. Nil means unknown; a zero price means free.
	ContextWindow  int      `json:"context_window,omitempty"`
	CostPerMInput  *float64 `json:"cost_per_m_input,omitempty"`
	CostPerMOutput *float64 `json:"cost_per_m_output,omitempty"`

	// Local metadata.
	SizeBytes *int64 `json:"size_bytes,omitempty"`
}

// Has reports whether the model has all the given features.
func (m Model) Has(features ...Feature) bool {
	for _, f := range features {
		if !slices.Contains(m.Features, f) {
			return false
		}
	}
	return true
}

// UseCase serves chat.
type UseCase struct {
	*usecase.Base[Model, Provider]
}

// New builds the chat use case over the providers that implement Provider. It lists their models before it returns.
func New(ctx context.Context, providers []provider.Provider) *UseCase {
	return &UseCase{usecase.NewBase(ctx, usecase.Spec[Model, Provider]{
		Name:      "chat",
		List:      Provider.ListChatModels,
		Prefer:    prefer,
		Scenarios: scenarios,
	}, providers)}
}

// prefer picks the cheaper model in the cloud and the smaller one locally.
func prefer(a, b Model) bool {
	if a.ProviderType == provider.Local {
		return usecase.LessKnown(a.SizeBytes, b.SizeBytes)
	}
	return usecase.LessKnown(a.CostPerMInput, b.CostPerMInput)
}

// Handle serves POST /v1/chat.
func (u *UseCase) Handle(w http.ResponseWriter, r *http.Request) error {
	var req Request
	if err := api.DecodeJSON(r, &req); err != nil {
		return err
	}
	if err := req.validate(); err != nil {
		return err
	}

	model, p, err := u.Resolve(req.Model)
	if err != nil {
		return err
	}
	req.Model = model.ID

	resp, err := p.Chat(r.Context(), &req)
	if err != nil {
		return fmt.Errorf("%s: %w", p.Name(), err)
	}
	api.WriteJSON(w, resp)
	return nil
}

// validate checks the request's required fields and enum values.
func (r *Request) validate() error {
	if len(r.Messages) == 0 {
		return api.NewError(http.StatusBadRequest, "messages is required")
	}

	if r.ReasoningEffort != nil {
		switch *r.ReasoningEffort {
		case ReasoningEffortNone, ReasoningEffortLow, ReasoningEffortMedium, ReasoningEffortHigh:
		default:
			return api.NewError(http.StatusBadRequest, "reasoning_effort must be one of: none, low, medium, high")
		}
	}

	for _, m := range r.Messages {
		switch m.Role {
		case RoleSystem, RoleUser, RoleAssistant, RoleTool:
		default:
			return api.NewError(http.StatusBadRequest, "message role must be one of: system, user, assistant, tool")
		}
	}
	return nil
}
