package exa

import (
	"context"
	"strings"

	"github.com/mltheuser/ai-router/router"
	"github.com/mltheuser/ai-router/router/search"
)

// searchTypes are Exa's search modes, served as search models. Fast comes
// first: search models rank equal, so the first listed is the default pick
// (see router.Spec.Prefer).
var searchTypes = []string{"fast", "instant", "auto", "deep-lite", "deep", "deep-reasoning"}

// --- Search wire types ---

type searchRequest struct {
	Query      string         `json:"query"`
	Type       string         `json:"type"`
	NumResults *int           `json:"numResults,omitempty"`
	Contents   searchContents `json:"contents"`
}

type searchContents struct {
	Highlights bool `json:"highlights"`
}

type searchResponse struct {
	Results []searchResult `json:"results"`
}

type searchResult struct {
	URL        string   `json:"url"`
	Title      string   `json:"title"`
	Highlights []string `json:"highlights"`
}

// ListSearchModels returns Exa's search modes. They are fixed; Exa has no
// endpoint that lists them.
func (p *Provider) ListSearchModels(_ context.Context) ([]search.Model, error) {
	models := make([]search.Model, len(searchTypes))
	for i, t := range searchTypes {
		models[i] = search.Model{ModelRef: router.NewModelRef(p, t)}
	}
	return models, nil
}

// Search searches in the mode the model names. A result's snippet is its
// highlights: the excerpts Exa selects as relevant to the query.
func (p *Provider) Search(ctx context.Context, req *search.Request) (*search.Response, error) {
	wireReq := searchRequest{
		Query:      req.Query,
		Type:       req.Model,
		NumResults: req.MaxResults,
		Contents:   searchContents{Highlights: true},
	}

	var wireResp searchResponse
	if err := p.client.Post(ctx, "/search", wireReq, &wireResp); err != nil {
		return nil, err
	}

	resp := &search.Response{Model: req.Model, Results: make([]search.Result, 0, len(wireResp.Results))}
	for _, r := range wireResp.Results {
		resp.Results = append(resp.Results, search.Result{
			URL:     r.URL,
			Title:   r.Title,
			Snippet: strings.Join(r.Highlights, "\n\n"),
		})
	}
	return resp, nil
}
