// Package search is the search use case: a query in, a ranked list of web
// pages out.
package search

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mltheuser/ai-router/httpx"
	"github.com/mltheuser/ai-router/router"
)

// Provider is implemented by every backend that serves web search.
type Provider interface {
	router.Provider

	// ListSearchModels returns the search models available at this provider.
	ListSearchModels(ctx context.Context) ([]Model, error)

	// Search returns the web pages that match the query. req.Model is the
	// bare model ID, as listed by ListSearchModels.
	Search(ctx context.Context, req *Request) (*Response, error)
}

// Model describes a search model at one provider: one way the provider can
// search, e.g. a fast mode or a deep research mode.
type Model struct {
	router.ModelRef
}

// UseCase serves web search.
type UseCase struct {
	*router.Base[Model, Provider]
}

// New builds the search use case over the providers that implement Provider.
func New(ctx context.Context, providers []router.Provider) *UseCase {
	return &UseCase{router.NewBase(ctx, router.Spec[Model, Provider]{
		Name:      "search",
		List:      Provider.ListSearchModels,
		Prefer:    prefer,
		Scenarios: scenarios,
	}, providers)}
}

// prefer ranks all models equal: they carry no metadata to rank by.
func prefer(_, _ Model) bool { return false }

// Handle serves POST /v1/search.
func (u *UseCase) Handle(w http.ResponseWriter, r *http.Request) error {
	var req Request
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return err
	}

	model, p, err := u.Resolve(req.Model)
	if err != nil {
		return err
	}
	req.Model = model.ID

	resp, err := p.Search(r.Context(), &req)
	if err != nil {
		return fmt.Errorf("%s: %w", p.Name(), err)
	}
	httpx.WriteJSON(w, resp)
	return nil
}
