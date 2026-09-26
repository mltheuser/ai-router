package openai

import (
	"context"
	"fmt"

	"github.com/mltheuser/ai-router/api"
)

type embedRequest struct {
	Model      string   `json:"model"`
	Input      []string `json:"input"`
	Dimensions *int     `json:"dimensions,omitempty"`
}

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

// Embed generates embeddings via the OpenAI embeddings endpoint.
func (p *Provider) Embed(ctx context.Context, req *api.EmbedRequest) (*api.EmbedResponse, error) {
	oReq := embedRequest{
		Model:      req.Model,
		Input:      req.Input,
		Dimensions: req.Dimensions,
	}

	var oResp embedResponse
	if err := p.client.post(ctx, "/embeddings", oReq, &oResp); err != nil {
		return nil, fmt.Errorf("openai embed: %w", err)
	}

	resp := &api.EmbedResponse{
		Object: oResp.Object,
		Model:  oResp.Model,
		Usage: api.EmbedUsage{
			PromptTokens: oResp.Usage.PromptTokens,
			TotalTokens:  oResp.Usage.TotalTokens,
		},
	}
	for _, d := range oResp.Data {
		resp.Data = append(resp.Data, api.EmbedData{
			Object:    d.Object,
			Embedding: d.Embedding,
			Index:     d.Index,
		})
	}

	return resp, nil
}
