package exa

import "github.com/mltheuser/ai-router/httpx"

const baseURL = "https://api.exa.ai"

// betaFeatures opts into Exa's dynamic highlights, which search uses and Exa
// still serves only behind this beta flag.
const betaFeatures = "dynamic-highlights-2026-08-28"

func newClient(apiKey string) *httpx.Client {
	return httpx.NewClient(baseURL,
		httpx.WithHeader("x-api-key", apiKey),
		httpx.WithHeader("Exa-Beta", betaFeatures),
	)
}
