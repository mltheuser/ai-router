// Package usecase holds the machinery every use case shares, written once:
// the per-use-case model catalog, model-string resolution, the model listing
// endpoint and the scenario test runner.
//
// Model lists are fetched once, when a use case is built at startup, and never
// change afterwards; restarting the server is how it picks up new models.
//
// A use case (chat, embedding, ...) lives in its own sub-package and owns its
// contract end to end: its model type, its request and response types, the
// provider interface a backend implements to serve it, its request handler
// and its test scenarios. It declares itself with a Spec and embeds the Base
// built from it, which implements everything in UseCase except Handle.
package usecase

import (
	"context"
	"net/http"
	"sort"

	"github.com/mltheuser/ai-router/provider"
)

// UseCase is what the server needs from every use case. The server derives
// the routes from Name, so every use case is served the same way:
//
//	POST /v1/<name>         Handle
//	GET  /v1/<name>/models  ListModels
type UseCase interface {
	// Name identifies the use case in routes, logs and test reports.
	Name() string

	// Handle serves one request. It writes the response on success; on
	// failure it writes nothing and returns the error for the server to write.
	Handle(w http.ResponseWriter, r *http.Request) error

	// ListModels writes the models that Handle accepts.
	ListModels(w http.ResponseWriter, r *http.Request) error

	// Test verifies one provider's implementation of the use case against
	// the models it serves; see TestRequest. It reports false if the provider
	// does not serve this use case.
	Test(ctx context.Context, req TestRequest) (Report, bool)
}

// Model is the constraint on a use case's model type: it embeds
// provider.ModelRef.
type Model interface {
	Ref() provider.ModelRef
}

// Spec declares a use case to the shared machinery.
type Spec[M Model, P provider.Provider] struct {
	// Name identifies the use case; see UseCase.Name.
	Name string

	// List fetches the use case's models from one provider. It is P's listing
	// method as a method expression, e.g. chat.Provider.ListChatModels.
	List func(P, context.Context) ([]M, error)

	// Prefer reports whether a is a better pick than b. Resolve uses it to
	// choose between providers serving the same model, Test to choose the
	// model a scenario runs against. a and b always share a provider type.
	Prefer func(a, b M) bool

	// Scenarios verify a provider's implementation end to end; see Test.
	Scenarios []Scenario[M]
}

// Base implements the shared part of UseCase from a Spec. A use case embeds
// it and adds Handle, which calls Resolve to pick the model and provider.
//
// A Base is read-only once built, so it is safe for concurrent use.
type Base[M Model, P provider.Provider] struct {
	spec      Spec[M, P]
	providers map[string]P   // the providers serving this use case, by name
	names     []string       // their names, sorted, for deterministic iteration
	models    map[string][]M // provider name → its model list
}

// NewBase builds the Base of the use case declared by spec. Of the given
// providers it serves exactly those that implement P, and lists their models
// before it returns.
func NewBase[M Model, P provider.Provider](ctx context.Context, spec Spec[M, P], providers []provider.Provider) *Base[M, P] {
	b := &Base[M, P]{spec: spec, providers: make(map[string]P)}
	for _, p := range providers {
		if up, ok := p.(P); ok {
			b.providers[p.Name()] = up
			b.names = append(b.names, p.Name())
		}
	}
	sort.Strings(b.names)
	b.models = b.listAll(ctx)
	return b
}

func (b *Base[M, P]) Name() string { return b.spec.Name }
