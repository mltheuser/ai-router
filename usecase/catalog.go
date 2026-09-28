package usecase

import (
	"cmp"
	"context"
	"log/slog"
	"net/http"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/mltheuser/ai-router/api"
	"github.com/mltheuser/ai-router/provider"
)

// ListModels writes the listed models, sorted by provider and ID.
// Optional query parameters narrow the list: type=cloud|local, and search,
// a case-insensitive substring of the model ID.
func (b *Base[M, P]) ListModels(w http.ResponseWriter, r *http.Request) error {
	providerType := provider.Type(r.URL.Query().Get("type"))
	if providerType != "" && providerType != provider.Cloud && providerType != provider.Local {
		return api.NewError(http.StatusBadRequest, "type must be 'cloud' or 'local'")
	}
	search := strings.ToLower(r.URL.Query().Get("search"))

	models := []M{} // encodes as [], never null
	b.mu.RLock()
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
	b.mu.RUnlock()
	slices.SortFunc(models, func(a, b M) int {
		return cmp.Or(cmp.Compare(a.Ref().Provider, b.Ref().Provider), cmp.Compare(a.Ref().ID, b.Ref().ID))
	})

	api.WriteJSON(w, struct {
		Data []M `json:"data"`
	}{models})
	return nil
}

func (b *Base[M, P]) Start(ctx context.Context) {
	b.Refresh(ctx)
	go func() {
		ticker := time.NewTicker(localTTL)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				b.refreshStale(ctx)
			}
		}
	}()
}

func (b *Base[M, P]) Refresh(ctx context.Context) {
	b.refreshEach(ctx, b.names)
}

// refreshStale re-lists the providers whose list is older than their TTL.
func (b *Base[M, P]) refreshStale(ctx context.Context) {
	var stale []string
	b.mu.RLock()
	for _, name := range b.names {
		ttl := cloudTTL
		if b.providers[name].Type() == provider.Local {
			ttl = localTTL
		}
		if time.Since(b.listedAt[name]) >= ttl {
			stale = append(stale, name)
		}
	}
	b.mu.RUnlock()
	b.refreshEach(ctx, stale)
}

// refreshEach re-lists the named providers in parallel and waits for all.
func (b *Base[M, P]) refreshEach(ctx context.Context, names []string) {
	var wg sync.WaitGroup
	for _, name := range names {
		wg.Go(func() { b.refresh(ctx, name) })
	}
	wg.Wait()
}

// refresh replaces one provider's model list. On failure the previous list
// stays in place and is retried on the next tick.
func (b *Base[M, P]) refresh(ctx context.Context, name string) {
	models, err := b.spec.List(b.providers[name], ctx)
	if err != nil {
		slog.Warn("Listing models failed", "use_case", b.spec.Name, "provider", name, "error", err)
		return
	}

	b.mu.Lock()
	b.models[name] = models
	b.listedAt[name] = time.Now()
	b.mu.Unlock()
	slog.Debug("Listed models", "use_case", b.spec.Name, "provider", name, "models", len(models))
}
