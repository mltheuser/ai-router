package anthropic

import "github.com/mltheuser/ai-router/httpx"

const baseURL = "https://api.anthropic.com/v1"

// anthropicVersion pins the API version contract via the required
// anthropic-version header. It is a version identifier, not a release date:
// 2023-06-01 is the current stable value and rarely changes.
const anthropicVersion = "2023-06-01"

func newClient(apiKey string) *httpx.Client {
	return httpx.NewClient(baseURL,
		httpx.WithHeader("x-api-key", apiKey),
		httpx.WithHeader("anthropic-version", anthropicVersion),
	)
}
