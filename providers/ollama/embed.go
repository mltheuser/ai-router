package ollama

import (
	"context"
	"fmt"

	"github.com/mltheuser/ai-router/api"
)

// embedRequest is the Ollama OpenAI-compatible embedding request.
type embedRequest struct {
	Model      string   `json:"model"`
	Input      []string `json:"input"`
	Dimensions     *int        `json:"dimensions,omitempty"`
}

// embedResponse is the Ollama OpenAI-compatible embedding response.
type embedResponse struct {
	Object string      `json:"object"`
	Data   []embedData `json:"data"`
	Model  string      `json:"model"`
	Usage  embedUsage  `json:"usage"`
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

// Embed generates embeddings via Ollama's OpenAI-compatible endpoint.
func (p *Provider) Embed(ctx context.Context, req *api.EmbedRequest) (*api.EmbedResponse, error) {
	provReq := embedRequest{
		Model:      req.Model,
		Input:      req.Input,
		Dimensions: req.Dimensions,
	}

	var provResp embedResponse
	if err := p.client.post(ctx, "/v1/embeddings", provReq, &provResp); err != nil {
		return nil, fmt.Errorf("ollama embed: %w", err)
	}

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
