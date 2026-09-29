package anthropic

import "github.com/mltheuser/ai-router/httpx"

const baseURL = "https://api.anthropic.com/v1"

// anthropicVersion is the API version sent in the required anthropic-version
// header. It names a version of the API contract, not a release date.
const anthropicVersion = "2023-06-01"

func newClient(apiKey string) *httpx.Client {
	return httpx.NewClient(baseURL,
		httpx.WithHeader("x-api-key", apiKey),
		httpx.WithHeader("anthropic-version", anthropicVersion),
	)
}
