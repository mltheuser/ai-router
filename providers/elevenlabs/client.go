package elevenlabs

import "github.com/mltheuser/ai-router/httpx"

const baseURL = "https://api.elevenlabs.io"

func newClient(apiKey string) *httpx.Client {
	return httpx.NewClient(baseURL, httpx.WithHeader("xi-api-key", apiKey))
}
