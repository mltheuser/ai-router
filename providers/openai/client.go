package openai

import "github.com/mltheuser/ai-router/httpx"

const baseURL = "https://api.openai.com/v1"

func newClient(apiKey string) *httpx.Client {
	return httpx.NewClient(baseURL, httpx.WithHeader("Authorization", "Bearer "+apiKey))
}
