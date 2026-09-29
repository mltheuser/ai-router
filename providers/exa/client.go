package exa

import "github.com/mltheuser/ai-router/httpx"

const baseURL = "https://api.exa.ai"

func newClient(apiKey string) *httpx.Client {
	return httpx.NewClient(baseURL, httpx.WithHeader("x-api-key", apiKey))
}
