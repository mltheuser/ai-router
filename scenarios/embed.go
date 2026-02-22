package scenarios

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"

	"github.com/mltheuser/ai-router/api"
)

func init() {
	Register(&EmbedBatchSimilarity{})
}

type EmbedBatchSimilarity struct{}

func (s *EmbedBatchSimilarity) Name() string {
	return "embed_batch_similarity"
}

func (s *EmbedBatchSimilarity) Description() string {
	return "Verifies batch embedding similarity and optional dimension control"
}

func (s *EmbedBatchSimilarity) RequiredCapabilities() []api.Capability {
	return []api.Capability{api.CapabilityEmbed}
}

func (s *EmbedBatchSimilarity) Run(ctx context.Context, baseURL string, modelID string) *api.ScenarioResult {
	url := fmt.Sprintf("%s/v1/embeddings", baseURL)
	client := http.DefaultClient
	result := api.NewResult()

	// Three inputs for similarity testing:
	// 1. Base text
	// 2. Similar text (1 word difference)
	// 3. Different text
	// Requesting a specific dimension lets us validate dimension control
	// on the same response.
	inputs := []string{
		"The quick brown fox jumps over the lazy dog.",
		"The quick brown fox jumps over the lazy cat.",
		"Planetary motion is governed by Kepler's laws.",
	}
	targetDim := 256

	embedResp, err := doEmbedRequest(ctx, client, url, modelID, inputs, &targetDim)
	if err != nil {
		result.Fail("batch embedding request", fmt.Sprintf("%v", err))
		return result
	}

	if len(embedResp.Data) != len(inputs) {
		result.Fail("batch embedding request", fmt.Sprintf("expected %d embeddings, got %d", len(inputs), len(embedResp.Data)))
		return result
	}

	checkDimensions(result, embedResp, targetDim)
	checkSimilarity(result, embedResp)

	return result
}

func checkDimensions(result *api.ScenarioResult, resp *api.EmbedResponse, targetDim int) {
	actualDim := len(resp.Data[0].Embedding)
	if actualDim != targetDim {
		result.Fail("dimension control", fmt.Sprintf("requested %d dimensions, got %d", targetDim, actualDim))
	} else {
		result.Pass("dimension control")
	}
}

func checkSimilarity(result *api.ScenarioResult, resp *api.EmbedResponse) {
	emb1 := resp.Data[0].Embedding
	emb2 := resp.Data[1].Embedding
	emb3 := resp.Data[2].Embedding

	sim12, err := cosineSimilarity(emb1, emb2)
	if err != nil {
		result.Fail("batch embedding similarity", fmt.Sprintf("failed to compute sim(1, 2): %v", err))
		return
	}
	sim13, err := cosineSimilarity(emb1, emb3)
	if err != nil {
		result.Fail("batch embedding similarity", fmt.Sprintf("failed to compute sim(1, 3): %v", err))
		return
	}

	if sim12 <= sim13 {
		result.Fail("batch embedding similarity", fmt.Sprintf("similar texts should be more similar than different texts (sim12=%f, sim13=%f)", sim12, sim13))
	} else {
		result.Pass("batch embedding similarity")
	}
}

// doEmbedRequest sends an embedding request and returns the decoded response.
func doEmbedRequest(ctx context.Context, client *http.Client, url, modelID string, inputs []string, dimensions *int) (*api.EmbedResponse, error) {
	reqBody := api.EmbedRequest{
		Model:      modelID,
		Input:      inputs,
		Dimensions: dimensions,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var embedResp api.EmbedResponse
	if err := json.NewDecoder(resp.Body).Decode(&embedResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &embedResp, nil
}

// cosineSimilarity computes cosine similarity between two vectors.
func cosineSimilarity(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vector lengths do not match: %d vs %d", len(a), len(b))
	}
	var dotProduct, normA, normB float64
	for i := 0; i < len(a); i++ {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0, fmt.Errorf("zero vector encountered")
	}
	return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB)), nil
}
