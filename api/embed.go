package api

// EmbedRequest represents an embedding request.
type EmbedRequest struct {
	Model      string   `json:"model"`
	Input      []string `json:"input"`
	Dimensions *int     `json:"dimensions,omitempty"`
}

// EmbedResponse represents an embedding response.
type EmbedResponse struct {
	Object string      `json:"object"`
	Data   []EmbedData `json:"data"`
	Model  string      `json:"model"`
	Usage  EmbedUsage  `json:"usage"`
}

// EmbedData represents a single embedding result.
type EmbedData struct {
	Object    string    `json:"object"`
	Embedding []float64 `json:"embedding"`
	Index     int       `json:"index"`
}

// EmbedUsage represents token usage for an embedding request.
type EmbedUsage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}
