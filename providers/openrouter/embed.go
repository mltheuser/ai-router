package openrouter

import (
	"context"
	"fmt"

	"github.com/mltheuser/ai-router/api"
)

// embedRequest is the OpenRouter-specific embedding request (OpenAI-compatible).
type embedRequest struct {
	Model      string   `json:"model"`
	Input      []string `json:"input"`
	Dimensions     *int     `json:"dimensions,omitempty"`
}

// embedResponse is the OpenRouter-specific embedding response (OpenAI-compatible).
type embedResponse struct {
	Object string          `json:"object"`
	Data   []embedData     `json:"data"`
	Model  string          `json:"model"`
	Usage  embedUsage      `json:"usage"`
}

type embedData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

type embedUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// Embed generates embeddings via OpenRouter (OpenAI-compatible endpoint).
func (p *Provider) Embed(ctx context.Context, req *api.EmbedRequest) (*api.EmbedResponse, error) {
	// Build provider-specific request
	// Build provider-specific request
	provReq := embedRequest{
		Model:      req.Model,
		Input:      req.Input,
		Dimensions: req.Dimensions,
	}

	var provResp embedResponse
	if err := p.client.post(ctx, "/embeddings", provReq, &provResp); err != nil {
		return nil, fmt.Errorf("openrouter embed: %w", err)
	}

	// Convert to unified response
	resp := &api.EmbedResponse{
		Object: provResp.Object,
		Model:  provResp.Model,
		Usage: api.EmbedUsage{
			PromptTokens: provResp.Usage.PromptTokens,
			TotalTokens:  provResp.Usage.TotalTokens,
		},
	}

	for _, d := range provResp.Data {
		resp.Data = append(resp.Data, api.EmbedData{
			Object:    d.Object,
			Embedding: d.Embedding,
			Index:     d.Index,
		})
	}

	return resp, nil
}
