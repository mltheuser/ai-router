package search

// Request is the body of POST /v1/search.
type Request struct {
	Model string `json:"model"`
	Query string `json:"query"`
	// MaxResults, if set, caps the number of results. The provider may
	// return fewer.
	MaxResults *int `json:"max_results,omitempty"`
}

// Response is the body of a successful POST /v1/search response. Results are
// ordered by relevance, most relevant first.
type Response struct {
	Model   string   `json:"model"`
	Results []Result `json:"results"`
}

// Result is one web page found for the query.
type Result struct {
	URL string `json:"url"`
	// Title is the page's title. It may be empty: not every page has one.
	Title string `json:"title"`
	// Snippet is what the page holds that matches the query: the excerpts
	// the provider selected as relevant. It is not necessarily short; its
	// length is up to the provider.
	Snippet string `json:"snippet"`
}
