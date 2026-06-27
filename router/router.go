package router

import (
	"fmt"
	"math"
	"strings"

	"github.com/mltheuser/ai-router/api"
	"github.com/mltheuser/ai-router/provider"
)

// Router resolves model_id:tag[@provider] to a specific provider instance.
type Router struct {
	catalog *ModelCatalog
}

// NewRouter creates a new router backed by the given catalog.
func NewRouter(catalog *ModelCatalog) *Router {
	return &Router{catalog: catalog}
}

// ResolveResult contains the resolved provider and cleaned model ID.
type ResolveResult struct {
	Provider  provider.Provider
	ModelID   string // model ID without the :tag and @provider parts
	ModelInfo api.ModelInfo
}

// Resolve parses a model string like "model_id:cloud@openrouter" and returns
// the appropriate provider and cleaned model ID.
func (r *Router) Resolve(modelStr string, capability api.Capability) (*ResolveResult, error) {
	modelID, tag, pinned, err := parseModelString(modelStr)
	if err != nil {
		return nil, err
	}

	// Get candidates from catalog
	candidates := r.catalog.GetProviders(modelID, tag)
	if len(candidates) == 0 {
		return nil, fmt.Errorf("%w: model '%s' not available with tag '%s'", api.ErrModelNotFound, modelID, tag)
	}

	// Filter by capability
	var capable []api.ModelInfo
	for _, c := range candidates {
		if c.HasCapability(capability) {
			capable = append(capable, c)
		}
	}
	if len(capable) == 0 {
		return nil, fmt.Errorf("%w: model '%s' does not support capability '%s'", api.ErrNotSupported, modelID, capability)
	}

	// If provider is pinned, filter to that provider
	if pinned != "" {
		var pinResult []api.ModelInfo
		for _, c := range capable {
			if c.Provider == pinned {
				pinResult = append(pinResult, c)
			}
		}
		if len(pinResult) == 0 {
			return nil, fmt.Errorf("%w: model '%s' not available at provider '%s'", api.ErrModelNotFound, modelID, pinned)
		}
		capable = pinResult
	}

	// Select best candidate
	selected, err := SelectBestCandidate(capable, tag)
	if err != nil {
		return nil, err
	}

	// Get the actual provider instance
	p := r.catalog.GetProvider(selected.Provider)
	if p == nil {
		return nil, fmt.Errorf("%w: provider '%s' not registered", api.ErrProviderUnavailable, selected.Provider)
	}

	return &ResolveResult{
		Provider:  p,
		ModelID:   modelID,
		ModelInfo: selected,
	}, nil
}

// parseModelString parses "model_id:tag[@provider]" into components.
func parseModelString(s string) (modelID string, tag api.ProviderType, pinned string, err error) {
	// Check for @provider suffix
	if idx := strings.LastIndex(s, "@"); idx != -1 {
		pinned = s[idx+1:]
		s = s[:idx]
	}

	// Check for :tag
	if idx := strings.LastIndex(s, ":"); idx != -1 {
		tagStr := s[idx+1:]
		modelID = s[:idx]

		switch tagStr {
		case "cloud":
			tag = api.ProviderTypeCloud
		case "local":
			tag = api.ProviderTypeLocal
		default:
			return "", "", "", fmt.Errorf("invalid tag '%s': must be 'cloud' or 'local'", tagStr)
		}
	} else {
		return "", "", "", fmt.Errorf("model '%s' missing required tag: use '%s:cloud' or '%s:local'", s, s, s)
	}

	if modelID == "" {
		return "", "", "", fmt.Errorf("empty model ID")
	}

	return modelID, tag, pinned, nil
}

// SelectBestCandidate picks the best provider from candidates.
// For cloud: cheapest by input cost. For local: first available.
func SelectBestCandidate(candidates []api.ModelInfo, tag api.ProviderType) (api.ModelInfo, error) {
	if len(candidates) == 0 {
		return api.ModelInfo{}, fmt.Errorf("no candidates provided")
	}

	if len(candidates) == 1 {
		return candidates[0], nil
	}

	if tag == api.ProviderTypeCloud {
		// Pick cheapest. Treat nil (unknown) as Infinity so we prefer known pricing.
		getCost := func(m api.ModelInfo) float64 {
			if m.CostPerMInput == nil {
				return math.Inf(1)
			}
			return *m.CostPerMInput
		}

		best := candidates[0]
		bestCost := getCost(best)

		for _, c := range candidates[1:] {
			cCost := getCost(c)
			if cCost < bestCost {
				best = c
				bestCost = cCost
			}
		}
		return best, nil
	}

	if tag == api.ProviderTypeLocal {
		// Pick smallest size. Treat nil (unknown) as MaxInt64 so we prefer known sizes.
		getSize := func(m api.ModelInfo) int64 {
			if m.SizeBytes == nil {
				return math.MaxInt64
			}
			return *m.SizeBytes
		}

		best := candidates[0]
		bestSize := getSize(best)

		for _, c := range candidates[1:] {
			cSize := getSize(c)
			if cSize < bestSize {
				best = c
				bestSize = cSize
			}
		}
		return best, nil
	}

	return api.ModelInfo{}, fmt.Errorf("invalid tag '%s': must be 'cloud' or 'local'", tag)
}
