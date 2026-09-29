// Package contents is the contents use case: web page URLs in, each page's
// content out. It owns the whole contract: the model type listed at
// GET /v1/contents/models, the request and response of POST /v1/contents,
// the Provider interface a backend implements to serve them, and the
// scenarios that verify an implementation.
package contents

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mltheuser/ai-router/httpx"
	"github.com/mltheuser/ai-router/router"
)

// Provider is implemented by every backend that serves page contents.
type Provider interface {
	router.Provider

	// ListContentsModels returns the contents models available at this
	// provider.
	ListContentsModels(ctx context.Context) ([]Model, error)

	// Contents returns the content of every requested page. req.Model is the
	// bare model ID, as listed by ListContentsModels.
	Contents(ctx context.Context, req *Request) (*Response, error)
}

// Model describes a contents model at one provider: one way the provider can
// load pages, e.g. from its cache or always fresh.
type Model struct {
	router.ModelRef
}

// UseCase serves page contents.
type UseCase struct {
	*router.Base[Model, Provider]
}

// New builds the contents use case over the providers that implement
// Provider. It lists their models before it returns.
func New(ctx context.Context, providers []router.Provider) *UseCase {
	return &UseCase{router.NewBase(ctx, router.Spec[Model, Provider]{
		Name:      "contents",
		List:      Provider.ListContentsModels,
		Prefer:    prefer,
		Scenarios: scenarios,
	}, providers)}
}

// prefer ranks no model above another: contents models carry no metadata to
// rank by, so the first one listed wins.
func prefer(_, _ Model) bool { return false }

// Handle serves POST /v1/contents.
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

	resp, err := p.Contents(r.Context(), &req)
	if err != nil {
		return fmt.Errorf("%s: %w", p.Name(), err)
	}
	httpx.WriteJSON(w, resp)
	return nil
}
