// Package embedding is the embedding use case: texts in, one vector per text
// out.
package embedding

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mltheuser/ai-router/httpx"
	"github.com/mltheuser/ai-router/router"
)

// Provider is implemented by every backend that serves embeddings.
type Provider interface {
	router.Provider

	// ListEmbeddingModels returns the embedding models available at this
	// provider.
	ListEmbeddingModels(ctx context.Context) ([]Model, error)

	// Embed returns one embedding per input text. req.Model is the bare model
	// ID, as listed by ListEmbeddingModels.
	Embed(ctx context.Context, req *Request) (*Response, error)
}

// Model describes an embedding model at one provider.
type Model struct {
	router.ModelRef

	// Cloud metadata. Nil means unknown; a zero price means free.
	ContextWindow int      `json:"context_window,omitempty"`
	CostPerMInput *float64 `json:"cost_per_m_input,omitempty"`

	// Local metadata.
	SizeBytes *int64 `json:"size_bytes,omitempty"`
}

// UseCase serves embeddings.
type UseCase struct {
	*router.Base[Model, Provider]
}

// New builds the embedding use case over the providers that implement
// Provider.
func New(ctx context.Context, providers []router.Provider) *UseCase {
	return &UseCase{router.NewBase(ctx, router.Spec[Model, Provider]{
		Name:      "embedding",
		List:      Provider.ListEmbeddingModels,
		Prefer:    prefer,
		Scenarios: scenarios,
	}, providers)}
}

// prefer picks the cheaper model in the cloud and the smaller one locally.
func prefer(a, b Model) bool {
	if a.ProviderType == router.Local {
		return router.LessKnown(a.SizeBytes, b.SizeBytes)
	}
	return router.LessKnown(a.CostPerMInput, b.CostPerMInput)
}

// Handle serves POST /v1/embedding.
func (u *UseCase) Handle(w http.ResponseWriter, r *http.Request) error {
	var req Request
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return err
	}
	if len(req.Input) == 0 {
		return httpx.NewError(http.StatusBadRequest, "input is required")
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
	httpx.WriteJSON(w, resp)
	return nil
}
