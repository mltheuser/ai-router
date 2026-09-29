package router

import (
	"context"
	"net/http"
	"sort"
)

// UseCase is what the server needs from every use case.
type UseCase interface {
	// Name identifies the use case in routes (/v1/<name>), logs and test
	// reports.
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
// ModelRef.
type Model interface {
	Ref() ModelRef
}

// Spec declares a use case to the shared machinery.
type Spec[M Model, P Provider] struct {
	// Name identifies the use case; see UseCase.Name.
	Name string

	// List fetches the use case's models from one provider. It is P's listing
	// method as a method expression, e.g. chat.Provider.ListChatModels.
	List func(P, context.Context) ([]M, error)

	// Prefer reports whether a is a better pick than b. Resolve uses it to
	// choose between providers serving the same model, Test to choose the
	// model a scenario runs against. a and b always share a provider type.
	// Among models it ranks equal, the one listed first wins.
	Prefer func(a, b M) bool

	// Scenarios verify a provider's implementation end to end; see Test.
	Scenarios []Scenario[M]
}

// Base implements the shared part of UseCase from a Spec. It is read-only once
// built, so it is safe for concurrent use.
type Base[M Model, P Provider] struct {
	spec      Spec[M, P]
	providers map[string]P   // the providers serving this use case, by name
	names     []string       // their names, sorted, for deterministic iteration
	models    map[string][]M // provider name → its model list
}

// NewBase builds the Base of the use case declared by spec. Of the given
// providers it serves exactly those that implement P, and lists their models
// before it returns.
func NewBase[M Model, P Provider](ctx context.Context, spec Spec[M, P], providers []Provider) *Base[M, P] {
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
