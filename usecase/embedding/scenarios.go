package embedding

import (
	"context"
	"fmt"
	"math"

	"github.com/mltheuser/ai-router/usecase"
)

var scenarios = []usecase.Scenario[Model]{
	{Name: "batch_similarity", Run: runBatchSimilarity},
}

// runBatchSimilarity verifies batch embedding, dimension control, and that
// similar texts embed closer together than different ones.
func runBatchSimilarity(ctx context.Context, url, model string, res *usecase.Result) {
	inputs := []string{
		"The quick brown fox jumps over the lazy dog.",
		"The quick brown fox jumps over the lazy cat.", // one word apart
		"Planetary motion is governed by Kepler's laws.",
	}
	dimensions := 256

	resp, err := usecase.PostJSON[Response](ctx, url, Request{Model: model, Input: inputs, Dimensions: &dimensions})
	if err != nil {
		res.Fail("batch embedding", err.Error())
		return
	}
	if len(resp.Data) != len(inputs) {
		res.Fail("batch embedding", fmt.Sprintf("expected %d embeddings, got %d", len(inputs), len(resp.Data)))
		return
	}
	res.Pass("batch embedding")

	if got := len(resp.Data[0].Embedding); got != dimensions {
		res.Fail("dimension control", fmt.Sprintf("requested %d dimensions, got %d", dimensions, got))
	} else {
		res.Pass("dimension control")
	}

	similar, err := cosineSimilarity(resp.Data[0].Embedding, resp.Data[1].Embedding)
	if err != nil {
		res.Fail("similarity", err.Error())
		return
	}
	different, err := cosineSimilarity(resp.Data[0].Embedding, resp.Data[2].Embedding)
	if err != nil {
		res.Fail("similarity", err.Error())
		return
	}
	if similar <= different {
		res.Fail("similarity", fmt.Sprintf("similar texts should embed closer than different ones (similar=%f, different=%f)", similar, different))
		return
	}
	res.Pass("similarity")
}

func cosineSimilarity(a, b []float64) (float64, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vector lengths differ: %d vs %d", len(a), len(b))
	}
	var dot, normA, normB float64
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA == 0 || normB == 0 {
		return 0, fmt.Errorf("zero vector")
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB)), nil
}
