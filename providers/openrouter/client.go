package openrouter

import "github.com/mltheuser/ai-router/httpx"

const baseURL = "https://openrouter.ai/api/v1"

func newClient(apiKey string) *httpx.Client {
	return httpx.NewClient(baseURL, httpx.WithHeader("Authorization", "Bearer "+apiKey))
}
