package search

import (
	"context"
	"fmt"
	"net/url"
	"strings"

	"github.com/mltheuser/ai-router/router"
)

var scenarios = []router.Scenario[Model]{
	{Name: "ranked_results", Run: runRankedResults},
}

// runRankedResults verifies that a search returns well-formed results, honors
// max_results, and ranks the page the query describes near the top. The
// query names RFC 2119, a document published in 1997 that never changes, so
// its page is a stable expected result.
func runRankedResults(ctx context.Context, endpoint string, m Model, res *router.Result) {
	model := m.Ref().Model
	maxResults := 5
	resp, err := router.PostJSON[Response](ctx, endpoint, Request{
		Model:      model,
		Query:      "RFC 2119 key words for use in RFCs to indicate requirement levels",
		MaxResults: &maxResults,
	})
	if err != nil {
		res.Fail("search", err.Error())
		return
	}
	if len(resp.Results) == 0 {
		res.Fail("search", "no results")
		return
	}
	res.Pass("search")

	if len(resp.Results) > maxResults {
		res.Fail("max_results", fmt.Sprintf("requested at most %d results, got %d", maxResults, len(resp.Results)))
	} else {
		res.Pass("max_results")
	}

	if err := checkFields(resp.Results); err != nil {
		res.Fail("result fields", err.Error())
	} else {
		res.Pass("result fields")
	}

	const top = 3
	for i, r := range resp.Results[:min(top, len(resp.Results))] {
		if strings.Contains(strings.ToLower(r.URL), "rfc2119") {
			res.Pass(fmt.Sprintf("relevant result (rank %d)", i+1))
			return
		}
	}
	res.Fail("relevant result", fmt.Sprintf("no page of RFC 2119 among the top %d results: %v", top, urls(resp.Results)))
}

// checkFields checks that every result has an absolute web URL and that the
// top result has a title and a snippet. Other results may lack a title: not
// every page has one.
func checkFields(results []Result) error {
	for i, r := range results {
		u, err := url.Parse(r.URL)
		if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
			return fmt.Errorf("result %d has no absolute http(s) URL: %q", i+1, r.URL)
		}
	}
	if results[0].Title == "" {
		return fmt.Errorf("top result has no title: %s", results[0].URL)
	}
	if results[0].Snippet == "" {
		return fmt.Errorf("top result has no snippet: %s", results[0].URL)
	}
	return nil
}

func urls(results []Result) []string {
	us := make([]string, len(results))
	for i, r := range results {
		us[i] = r.URL
	}
	return us
}
