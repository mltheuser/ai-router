package embedding

// This file holds the wire types of the embedding API: what clients send to
// POST /v1/embedding and receive back. The SDKs mirror these types.

// Request is the body of POST /v1/embedding.
type Request struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
	// Dimensions, if set, asks for vectors of this size. Not every model
	// supports it.
	Dimensions *int `json:"dimensions,omitempty"`
}

// Response is the body of a successful POST /v1/embedding response.
type Response struct {
	Model string      `json:"model"`
	Data  []Embedding `json:"data"`
	Usage Usage       `json:"usage"`
}

// Embedding is the vector of the input text at Index.
type Embedding struct {
	Index     int       `json:"index"`
	Embedding []float64 `json:"embedding"`
}

// Usage reports the tokens a request consumed.
type Usage struct {
	PromptTokens int `json:"prompt_tokens"`
	TotalTokens  int `json:"total_tokens"`
}
