// Package embedding is the embedding use case: texts in, one vector per text
// out. It owns the whole contract: the model type listed at
// GET /v1/embedding/models, the request and response of POST /v1/embedding,
// the Provider interface a backend implements to serve them, and the
// scenarios that verify an implementation.
package embedding

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mltheuser/ai-router/api"
	"github.com/mltheuser/ai-router/provider"
	"github.com/mltheuser/ai-router/usecase"
)

// Provider is implemented by every backend that serves embeddings.
type Provider interface {
	provider.Provider

	// ListEmbeddingModels returns the embedding models available at this
	// provider.
	ListEmbeddingModels(ctx context.Context) ([]Model, error)

	// Embed returns one embedding per input text. req.Model is the bare model
	// ID, as listed by ListEmbeddingModels.
	Embed(ctx context.Context, req *Request) (*Response, error)
}

// Model describes an embedding model at one provider.
type Model struct {
	provider.ModelRef

	// Cloud metadata. Nil means unknown; a zero price means free.
	ContextWindow int      `json:"context_window,omitempty"`
	CostPerMInput *float64 `json:"cost_per_m_input,omitempty"`

	// Local metadata.
	SizeBytes *int64 `json:"size_bytes,omitempty"`
}

// Request is the body of POST /v1/embedding.
type Request struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
	// Dimensions, if set, asks for vectors of this size. Not every model
	// supports it.
	Dimensions *int `json:"dimensions,omitempty"`
}

// Response is the body of a successful POST /v1/embedding response.
type Response struct {
	Model string      `json:"model"`
	Data  []Embedding `json:"data"`
	Usage Usage       `json:"usage"`
}

// Embedding is the vector of the input text at Index.
type Embedding struct {
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

// Usage reports the tokens a request consumed.
type Usage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// UseCase serves embeddings.
type UseCase struct {
	*usecase.Base[Model, Provider]
}

// New builds the embedding use case over the providers that implement
// Provider.
func New(providers []provider.Provider) *UseCase {
	return &UseCase{usecase.NewBase(usecase.Spec[Model, Provider]{
		Name:      "embedding",
		List:      Provider.ListEmbeddingModels,
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

// Handle serves POST /v1/embedding.
func (u *UseCase) Handle(w http.ResponseWriter, r *http.Request) error {
	var req Request
	if err := api.DecodeJSON(r, &req); err != nil {
		return err
	}
	if len(req.Input) == 0 {
		return api.NewError(http.StatusBadRequest, "input is required")
	}

	model, p, err := u.Resolve(req.Model)
	if err != nil {
		return err
	}
	req.Model = model.ID

	resp, err := p.Embed(r.Context(), &req)
	if err != nil {
		return fmt.Errorf("%s: %w", p.Name(), err)
	}
	api.WriteJSON(w, resp)
	return nil
}
