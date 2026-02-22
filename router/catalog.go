package router

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/mltheuser/ai-router/api"
	"github.com/mltheuser/ai-router/provider"
)

// ModelCatalog maintains a cached, thread-safe catalog of available models across all providers.
type ModelCatalog struct {
	mu          sync.RWMutex
	models      map[string][]api.ModelInfo // modelID → entries from all providers
	providers   map[string]provider.Provider
	lastRefresh map[string]time.Time

	cloudTTL time.Duration
	localTTL time.Duration
	logger   *slog.Logger
}

// NewModelCatalog creates a new catalog with the given providers.
func NewModelCatalog(logger *slog.Logger, providers []provider.Provider) *ModelCatalog {
	providerMap := make(map[string]provider.Provider, len(providers))
	for _, p := range providers {
		providerMap[p.Name()] = p
	}

	return &ModelCatalog{
		models:      make(map[string][]api.ModelInfo),
		providers:   providerMap,
		lastRefresh: make(map[string]time.Time),
		cloudTTL:    30 * time.Minute,
		localTTL:    30 * time.Second,
		logger:      logger,
	}
}

// Initialize polls all providers in parallel and builds the initial catalog.
// It blocks until the initial poll is complete.
func (c *ModelCatalog) Initialize(ctx context.Context) error {
	// Providers map is read-only, iterate directly
	var wg sync.WaitGroup
	errCount := 0
	var errMu sync.Mutex

	for name, p := range c.providers {
		wg.Add(1)
		go func(name string, p provider.Provider) {
			defer wg.Done()
			if err := c.refreshProvider(ctx, name, p); err != nil {
				errMu.Lock()
				errCount++
				errMu.Unlock()
				c.logger.Warn("Failed to poll provider", "provider", name, "error", err)
			}
		}(name, p)
	}
	wg.Wait()

	c.mu.RLock()
	totalModels := len(c.models)
	c.mu.RUnlock()

	if totalModels == 0 && errCount == len(c.providers) && len(c.providers) > 0 {
		return fmt.Errorf("all providers failed to initialize")
	}

	return nil
}

// StartBackgroundRefresh begins the background refresh loop.
// It should be called after Initialize.
func (c *ModelCatalog) StartBackgroundRefresh(ctx context.Context) {
	go c.backgroundRefreshLoop(ctx)
}

func (c *ModelCatalog) backgroundRefreshLoop(ctx context.Context) {
	ticker := time.NewTicker(c.localTTL)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Providers map is immutable, iterate directly
			for name, p := range c.providers {
				ttl := c.cloudTTL
				if p.Type() == api.ProviderTypeLocal {
					ttl = c.localTTL
				}

				c.mu.RLock()
				lastRefresh := c.lastRefresh[name]
				c.mu.RUnlock()

				if time.Since(lastRefresh) >= ttl {
					go c.refreshProvider(ctx, name, p)
				}
			}
		}
	}
}

// RefreshAll forces a refresh of all providers.
func (c *ModelCatalog) RefreshAll(ctx context.Context) {
	// Providers map is immutable, iterate directly
	var wg sync.WaitGroup
	for name, p := range c.providers {
		wg.Add(1)
		go func(name string, p provider.Provider) {
			defer wg.Done()
			c.refreshProvider(ctx, name, p)
		}(name, p)
	}
	wg.Wait()
}

// refreshProvider fetches models from a single provider and updates the catalog atomically.
func (c *ModelCatalog) refreshProvider(ctx context.Context, name string, p provider.Provider) error {
	models, err := p.ListModels(ctx)
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	// Remove old entries from this provider
	for modelID, entries := range c.models {
		filtered := entries[:0]
		for _, e := range entries {
			if e.Provider != name {
				filtered = append(filtered, e)
			}
		}
		if len(filtered) == 0 {
			delete(c.models, modelID)
		} else {
			c.models[modelID] = filtered
		}
	}

	// Add new entries
	for _, m := range models {
		c.models[m.ID] = append(c.models[m.ID], m)
	}

	c.lastRefresh[name] = time.Now()
	c.logger.Info("Refreshed provider", "provider", name, "models", len(models))
	return nil
}

// GetProviders returns all providers for a model ID filtered by provider type.
func (c *ModelCatalog) GetProviders(modelID string, providerType api.ProviderType) []api.ModelInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entries := c.models[modelID]
	var result []api.ModelInfo
	for _, e := range entries {
		if e.ProviderType == providerType {
			result = append(result, e)
		}
	}
	return result
}

// AllModels returns all models in the catalog, optionally filtered.
func (c *ModelCatalog) AllModels(providerType *api.ProviderType, capability *api.Capability) []api.ModelInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []api.ModelInfo
	seen := make(map[string]bool)

	for _, entries := range c.models {
		for _, e := range entries {
			if providerType != nil && e.ProviderType != *providerType {
				continue
			}
			if capability != nil && !e.HasCapability(*capability) {
				continue
			}
			key := e.ID + "|" + e.Provider
			if !seen[key] {
				seen[key] = true
				result = append(result, e)
			}
		}
	}
	return result
}

// GetProvider returns the registered provider instance by name.
func (c *ModelCatalog) GetProvider(name string) provider.Provider {
	return c.providers[name]
}
