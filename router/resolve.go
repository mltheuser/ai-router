package router

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mltheuser/ai-router/httpx"
)

var (
	errInvalidModel  = httpx.NewError(http.StatusBadRequest, "invalid model string")
	errModelNotFound = httpx.NewError(http.StatusNotFound, "model not found")
)

// query is a parsed request model string "id:provider_type[@provider]".
type query struct {
	id           string
	providerType ProviderType
	provider     string // empty unless the request pins a provider
}

// parseModel parses a request model string. The provider type tag is
// required; the @provider suffix is optional.
func parseModel(s string) (query, error) {
	var q query
	if idx := strings.LastIndex(s, "@"); idx != -1 {
		q.provider = s[idx+1:]
		s = s[:idx]
	}

	idx := strings.LastIndex(s, ":")
	if idx == -1 {
		return query{}, fmt.Errorf("%w: model '%s' missing required tag: use '%s:cloud' or '%s:local'", errInvalidModel, s, s, s)
	}
	q.id = s[:idx]
	switch tag := ProviderType(s[idx+1:]); tag {
	case Cloud, Local:
		q.providerType = tag
	default:
		return query{}, fmt.Errorf("%w: invalid tag '%s': must be 'cloud' or 'local'", errInvalidModel, tag)
	}

	if q.id == "" {
		return query{}, fmt.Errorf("%w: empty model ID", errInvalidModel)
	}
	return q, nil
}

// Resolve picks the model and provider a request's model string addresses:
// among the listed models with the string's ID and provider type, restricted
// to the pinned provider if there is one, the one Spec.Prefer ranks best.
func (b *Base[M, P]) Resolve(model string) (M, P, error) {
	var (
		best  M
		bestP P
		found bool
	)
	q, err := parseModel(model)
	if err != nil {
		return best, bestP, err
	}

	for _, name := range b.names {
		if q.provider != "" && q.provider != name {
			continue
		}
		for _, m := range b.models[name] {
			if ref := m.Ref(); ref.ID != q.id || ref.ProviderType != q.providerType {
				continue
			}
			if !found || b.spec.Prefer(m, best) {
				best, bestP, found = m, b.providers[name], true
			}
		}
	}

	if !found {
		return best, bestP, fmt.Errorf("%w: no %s model matches '%s'; see GET /v1/%s/models",
			errModelNotFound, b.spec.Name, model, b.spec.Name)
	}
	return best, bestP, nil
}
