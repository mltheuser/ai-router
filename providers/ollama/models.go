package ollama

import (
	"context"
	"fmt"

	"github.com/mltheuser/ai-router/api"
)

// ollamaTagsResponse is the response from GET /api/tags.
type ollamaTagsResponse struct {
	Models []ollamaModelEntry `json:"models"`
}

type ollamaModelEntry struct {
	Name       string             `json:"name"`
	Model      string             `json:"model"`
	ModifiedAt string             `json:"modified_at"`
	Size       int64              `json:"size"`
	Digest     string             `json:"digest"`
	Details    ollamaModelDetails `json:"details"`
}

type ollamaModelDetails struct {
	ParentModel   string   `json:"parent_model"`
	Format        string   `json:"format"`
	Family        string   `json:"family"`
	Families      []string `json:"families"`
	ParameterSize string   `json:"parameter_size"`
	QuantLevel    string   `json:"quantization_level"`
}

// ollamaShowRequest is the request for POST /api/show.
type ollamaShowRequest struct {
	Name string `json:"name"`
}

// ollamaShowResponse captures the fields we need from /api/show.
type ollamaShowResponse struct {
	Capabilities []string `json:"capabilities"`
}

// ListModels fetches installed models from Ollama and enriches with capabilities via /api/show.
func (p *Provider) ListModels(ctx context.Context) ([]api.ModelInfo, error) {
	// Step 1: Get list of installed models
	var tagsResp ollamaTagsResponse
	if err := p.client.get(ctx, "/api/tags", &tagsResp); err != nil {
		return nil, fmt.Errorf("listing ollama models: %w", err)
	}

	// Step 2: For each model, call /api/show to get capabilities
	var models []api.ModelInfo
	for _, m := range tagsResp.Models {
		caps, err := p.getCapabilities(ctx, m.Name)
		if err != nil {
			// If we can't get capabilities, default to chat
			caps = []api.Capability{api.CapabilityChat}
		}

		models = append(models, api.ModelInfo{
			ID:            m.Name,
			Provider:      p.Name(),
			ProviderType:  api.ProviderTypeLocal,
			Capabilities:  caps,
			SizeBytes:     int64Ptr(m.Size),
			// Local models are free
			CostPerMInput:  float64Ptr(0),
			CostPerMOutput: float64Ptr(0),
		})
	}

	return models, nil
}

// getCapabilities calls /api/show for a model and maps Ollama capabilities to our types.
func (p *Provider) getCapabilities(ctx context.Context, modelName string) ([]api.Capability, error) {
	var showResp ollamaShowResponse
	if err := p.client.post(ctx, "/api/show", ollamaShowRequest{Name: modelName}, &showResp); err != nil {
		return nil, err
	}

	var caps []api.Capability
	for _, c := range showResp.Capabilities {
		switch c {
		case "embedding":
			caps = append(caps, api.CapabilityEmbed)
		case "completion":
			caps = append(caps, api.CapabilityChat)
			caps = append(caps, api.CapabilityStructuredOutput)
		case "tools":
			caps = append(caps, api.CapabilityTools)
		case "vision":
			caps = append(caps, api.CapabilityVision)
		case "thinking":
			caps = append(caps, api.CapabilityReasoning)
		}
	}

	return caps, nil
}

func float64Ptr(v float64) *float64 {
	return &v
}
func int64Ptr(v int64) *int64 {
	return &v
}
