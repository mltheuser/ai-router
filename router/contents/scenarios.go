package contents

import (
	"context"
	"fmt"
	"strings"

	"github.com/mltheuser/ai-router/router"
)

var scenarios = []router.Scenario[Model]{
	{Name: "batch_with_failure", Run: runBatchWithFailure},
}

const (
	// rfc2119 is a page that never changes: RFC 2119, published in 1997.
	rfc2119 = "https://www.rfc-editor.org/rfc/rfc2119"
	// unreachable can never load: the .invalid top-level domain is reserved
	// by RFC 2606 never to resolve.
	unreachable = "https://nonexistent.invalid/"
)

// runBatchWithFailure verifies a batch that mixes a loadable page, a page
// that cannot load, and a repeat of the first: one result per requested URL
// in request order, the page's text, the failure reported on its own entry
// without failing the request, and the repeat answered like the original.
func runBatchWithFailure(ctx context.Context, endpoint, model string, res *router.Result) {
	urls := []string{rfc2119, unreachable, rfc2119}
	resp, err := router.PostJSON[Response](ctx, endpoint, Request{Model: model, URLs: urls})
	if err != nil {
		res.Fail("batch", err.Error())
		return
	}
	if len(resp.Results) != len(urls) {
		res.Fail("batch", fmt.Sprintf("requested %d URLs, got %d results", len(urls), len(resp.Results)))
		return
	}
	res.Pass("batch")

	for i, r := range resp.Results {
		if r.URL != urls[i] {
			res.Fail("request order", fmt.Sprintf("result %d is for %q, requested %q", i+1, r.URL, urls[i]))
			return
		}
	}
	res.Pass("request order")

	page := resp.Results[0]
	switch {
	case page.Error != "":
		res.Fail("page text", fmt.Sprintf("%s failed to load: %s", rfc2119, page.Error))
	case page.Title == "":
		res.Fail("page text", "the page has no title")
	case !strings.Contains(page.Text, "SHALL NOT"):
		// The phrase is defined in section 2 of the RFC.
		res.Fail("page text", fmt.Sprintf("text lacks the RFC's phrase 'SHALL NOT' (%d characters)", len(page.Text)))
	default:
		res.Pass("page text")
	}

	if failed := resp.Results[1]; failed.Error == "" || failed.Text != "" {
		res.Fail("failed page reported", fmt.Sprintf("expected an error and no text for %s, got error %q and %d characters of text",
			unreachable, failed.Error, len(failed.Text)))
	} else {
		res.Pass("failed page reported")
	}

	if repeat := resp.Results[2]; repeat != page {
		res.Fail("repeated url", fmt.Sprintf("the repeat of %s differs from the original (error %q, %d vs %d characters of text)",
			rfc2119, repeat.Error, len(repeat.Text), len(page.Text)))
	} else {
		res.Pass("repeated url")
	}
}
