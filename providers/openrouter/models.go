package openrouter

import (
	"context"
	"fmt"
	"strconv"

	"github.com/mltheuser/ai-router/api"
)

// openRouterModelsResponse is the raw response from GET /api/v1/models.
type openRouterModelsResponse struct {
	Data []openRouterModel `json:"data"`
}

type openRouterModel struct {
	ID                  string                 `json:"id"`
	Name                string                 `json:"name"`
	Created             int64                  `json:"created"`
	ContextLength       int                    `json:"context_length"`
	Architecture        openRouterArchitecture `json:"architecture"`
	Pricing             openRouterPricing      `json:"pricing"`
	Description         string                 `json:"description"`
	SupportedParameters []string               `json:"supported_parameters"`
}

type openRouterArchitecture struct {
	Modality         string   `json:"modality"`
	InputModalities  []string `json:"input_modalities"`
	OutputModalities []string `json:"output_modalities"`
	Tokenizer        string   `json:"tokenizer"`
}

type openRouterPricing struct {
	Prompt     string `json:"prompt"`
	Completion string `json:"completion"`
	Request    string `json:"request"`
	Image      string `json:"image"`
}

// ListModels fetches all models from OpenRouter and converts to unified ModelInfo.
// OpenRouter serves chat/completion models at /models and embedding models at
// /embeddings/models — we query both and merge the results.
func (p *Provider) ListModels(ctx context.Context) ([]api.ModelInfo, error) {
	// Fetch chat/completion models
	var chatResp openRouterModelsResponse
	if err := p.client.get(ctx, "/models", &chatResp); err != nil {
		return nil, fmt.Errorf("listing openrouter chat models: %w", err)
	}

	// Fetch embedding models from the separate endpoint
	var embedResp openRouterModelsResponse
	if err := p.client.get(ctx, "/embeddings/models", &embedResp); err != nil {
		// Non-fatal: we can still serve chat models if embedding listing fails
		embedResp.Data = nil
	}

	var models []api.ModelInfo

	// Process chat models
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, m := range chatResp.Data {
		info := convertModel(m, p.Name())
		models = append(models, info)

		// check for reasoning effort support
		for _, param := range m.SupportedParameters {
			if param == "reasoning_effort" {
				p.supportsReasoningEffort[m.ID] = true
				break
			}
		}
	}

	// Process embedding models
	for _, m := range embedResp.Data {
		models = append(models, convertModel(m, p.Name()))
	}

	return models, nil
}

// convertModel converts an OpenRouter model to a unified ModelInfo.
func convertModel(m openRouterModel, providerName string) api.ModelInfo {
	info := api.ModelInfo{
		ID:            m.ID,
		Provider:      providerName,
		ProviderType:  api.ProviderTypeCloud,
		Capabilities:  inferCapabilities(m),
		ContextWindow: m.ContextLength,
	}

	// Parse pricing: OpenRouter returns cost per token as a string.
	// Convert to cost per million tokens.
	if promptCost, err := strconv.ParseFloat(m.Pricing.Prompt, 64); err == nil {
		c := promptCost * 1_000_000
		info.CostPerMInput = &c
	}
	if completionCost, err := strconv.ParseFloat(m.Pricing.Completion, 64); err == nil {
		c := completionCost * 1_000_000
		info.CostPerMOutput = &c
	}

	return info
}

// inferCapabilities determines model capabilities from OpenRouter's architecture metadata.
func inferCapabilities(m openRouterModel) []api.Capability {
	var caps []api.Capability

	hasEmbedOutput := false
	hasChatOutput := false

	for _, out := range m.Architecture.OutputModalities {
		switch out {
		case "embeddings":
			hasEmbedOutput = true
		case "text":
			hasChatOutput = true
		}
	}

	if hasEmbedOutput {
		caps = append(caps, api.CapabilityEmbed)
	}
	if hasChatOutput {
		caps = append(caps, api.CapabilityChat)

		// Check for vision support from input modalities
		for _, inp := range m.Architecture.InputModalities {
			if inp == "image" {
				caps = append(caps, api.CapabilityVision)
				break
			}
		}

		// Check for structured output and reasoning support
		for _, p := range m.SupportedParameters {
			switch p {
			case "structured_outputs":
				caps = append(caps, api.CapabilityStructuredOutput)
			case "reasoning":
				caps = append(caps, api.CapabilityReasoning)
			case "tools":
				caps = append(caps, api.CapabilityTools)
			}
		}
	}

	return caps
}
