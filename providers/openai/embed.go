package openai

import (
	"context"

	"github.com/mltheuser/ai-router/usecase/embedding"
)

// --- Embedding wire types ---

type embedRequest struct {
	Model      string   `json:"model"`
	Input      []string `json:"input"`
	Dimensions *int     `json:"dimensions,omitempty"`
}

type embedResponse struct {
	Model string      `json:"model"`
	Data  []embedData `json:"data"`
	Usage embedUsage  `json:"usage"`
}

type embedData struct {
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

type embedUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}

// Embed generates embeddings via the OpenAI embeddings endpoint.
func (p *Provider) Embed(ctx context.Context, req *embedding.Request) (*embedding.Response, error) {
	wireReq := embedRequest{
		Model:      req.Model,
		Input:      req.Input,
		Dimensions: req.Dimensions,
	}

	var wireResp embedResponse
	if err := p.client.Post(ctx, "/embeddings", wireReq, &wireResp); err != nil {
		return nil, err
	}

	resp := &embedding.Response{
		Model: wireResp.Model,
		Usage: embedding.Usage{
			PromptTokens: wireResp.Usage.PromptTokens,
			TotalTokens:  wireResp.Usage.TotalTokens,
		},
	}
	for _, d := range wireResp.Data {
		resp.Data = append(resp.Data, embedding.Embedding{Index: d.Index, Embedding: d.Embedding})
	}
	return resp, nil
}
