package contents

// This file holds the wire types of the contents API: what clients send to
// POST /v1/contents and receive back. The SDKs mirror these types.

// Request is the body of POST /v1/contents.
type Request struct {
	Model string   `json:"model"`
	URLs  []string `json:"urls"`
}

// Response is the body of a successful POST /v1/contents response: one
// Result per requested URL, in request order.
type Response struct {
	Model   string   `json:"model"`
	Results []Result `json:"results"`
}

// Result is the content of one requested page. A page that could not be
// loaded has Error set, the provider's reason passed through, and no Title or
// Text; the request as a whole still succeeds.
type Result struct {
	// URL is the page's URL as requested.
	URL   string `json:"url"`
	Title string `json:"title,omitempty"`
	// Text is the page's content as markdown.
	Text  string `json:"text,omitempty"`
	Error string `json:"error,omitempty"`
}
