package exa

import (
	"context"
	"fmt"

	"github.com/mltheuser/ai-router/router"
	"github.com/mltheuser/ai-router/router/contents"
)

// contentsModel is the one contents model: Exa's default freshness, which
// serves a page from Exa's cache when it has one and loads it otherwise.
const contentsModel = "auto"

// --- Contents wire types ---

type contentsRequest struct {
	URLs []string `json:"urls"`
	Text bool     `json:"text"`
}

// contentsResponse holds the loaded pages in Results and the outcome for
// every requested URL in Statuses. Both identify a page by its ID, which is
// the URL as requested.
type contentsResponse struct {
	Results  []contentsResult `json:"results"`
	Statuses []contentsStatus `json:"statuses"`
}

type contentsResult struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Text  string `json:"text"`
}

type contentsStatus struct {
	ID     string         `json:"id"`
	Status string         `json:"status"`
	Error  *contentsError `json:"error"`
}

type contentsError struct {
	Tag            string `json:"tag"`
	HTTPStatusCode *int   `json:"httpStatusCode"`
}

// ListContentsModels returns the one contents model.
func (p *Provider) ListContentsModels(_ context.Context) ([]contents.Model, error) {
	return []contents.Model{{ModelRef: router.NewModelRef(p, contentsModel)}}, nil
}

// Contents loads the requested pages as markdown, merging Exa's two lists,
// loaded pages and per-URL statuses, into the shared results.
func (p *Provider) Contents(ctx context.Context, req *contents.Request) (*contents.Response, error) {
	var wireResp contentsResponse
	if err := p.client.Post(ctx, "/contents", contentsRequest{URLs: req.URLs, Text: true}, &wireResp); err != nil {
		return nil, err
	}

	// Exa returns a page requested twice only once, so pages are looked up
	// by URL rather than matched by position.
	pages := make(map[string]contentsResult, len(wireResp.Results))
	for _, r := range wireResp.Results {
		pages[r.ID] = r
	}
	statuses := make(map[string]contentsStatus, len(wireResp.Statuses))
	for _, s := range wireResp.Statuses {
		statuses[s.ID] = s
	}

	resp := &contents.Response{Model: req.Model, Results: make([]contents.Result, 0, len(req.URLs))}
	for _, u := range req.URLs {
		if page, ok := pages[u]; ok {
			resp.Results = append(resp.Results, contents.Result{URL: u, Title: page.Title, Text: page.Text})
			continue
		}
		resp.Results = append(resp.Results, contents.Result{URL: u, Error: describeFailure(statuses[u])})
	}
	return resp, nil
}

// describeFailure explains why a page did not load, from its status.
func describeFailure(s contentsStatus) string {
	switch {
	case s.Error != nil && s.Error.HTTPStatusCode != nil:
		return fmt.Sprintf("%s (status %d)", s.Error.Tag, *s.Error.HTTPStatusCode)
	case s.Error != nil && s.Error.Tag != "":
		return s.Error.Tag
	case s.Status != "":
		return s.Status
	default:
		return "no status reported"
	}
}
