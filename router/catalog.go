package router

import (
	"cmp"
	"context"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"sync"

	"github.com/mltheuser/ai-router/httpx"
)

// ListModels writes the listed models, sorted by provider and ID.
func (b *Base[M, P]) ListModels(w http.ResponseWriter, r *http.Request) error {
	providerType := ProviderType(r.URL.Query().Get("type"))
	if providerType != "" && providerType != Cloud && providerType != Local {
		return httpx.NewError(http.StatusBadRequest, "type must be 'cloud' or 'local'")
	}
	search := strings.ToLower(r.URL.Query().Get("search"))

	models := []M{} // encodes as [], never null
	for _, name := range b.names {
		for _, m := range b.models[name] {
			ref := m.Ref()
			if providerType != "" && ref.ProviderType != providerType {
				continue
			}
			if search != "" && !strings.Contains(strings.ToLower(ref.ID), search) {
				continue
			}
			models = append(models, m)
		}
	}
	slices.SortFunc(models, func(a, b M) int {
		return cmp.Or(cmp.Compare(a.Ref().Provider, b.Ref().Provider), cmp.Compare(a.Ref().ID, b.Ref().ID))
	})

	httpx.WriteJSON(w, struct {
		Data []M `json:"data"`
	}{models})
	return nil
}

// listAll lists the models of every provider in parallel. A provider whose
// listing fails serves no models until the server restarts.
func (b *Base[M, P]) listAll(ctx context.Context) map[string][]M {
	lists := make([][]M, len(b.names))
	var wg sync.WaitGroup
	for i, name := range b.names {
		wg.Go(func() {
			models, err := b.spec.List(b.providers[name], ctx)
			if err != nil {
				slog.Warn("Listing models failed", "use_case", b.spec.Name, "provider", name, "error", err)
				return
			}
			slog.Info("Listed models", "use_case", b.spec.Name, "provider", name, "models", len(models))
			lists[i] = models
		})
	}
	wg.Wait()

	models := make(map[string][]M, len(b.names))
	for i, name := range b.names {
		models[name] = lists[i]
	}
	return models
}
