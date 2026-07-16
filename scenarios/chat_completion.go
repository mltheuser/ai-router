// Package scenarios holds the self-contained end-to-end test scenarios run by
// the server's /v1/test endpoint to verify live provider capabilities.
package scenarios

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mltheuser/ai-router/api"
)

func init() {
	Register(&chatMultiStep{})
}

// brindlemarkGuide is a large (>4096 token) original document
// about an invented nation. It is sent verbatim on both turns of the scenario so
// the repeated prefix can trigger a provider-side prompt cache read on turn 2.
// The invented proper nouns (capital "Velmoria", river "Quillsong") cannot be
// answered from training data, making recall meaningful and substring-checkable.
//
//go:embed resources/brindlemark_guide.md
var brindlemarkGuide string

type chatMultiStep struct{}

func (s *chatMultiStep) Name() string {
	return "chat_multi_step"
}

func (s *chatMultiStep) Description() string {
	return "Verifies multi-step conversation recall over a large document and observes prompt-cache reads"
}

func (s *chatMultiStep) RequiredCapabilities() []api.Capability {
	return []api.Capability{api.CapabilityChat}
}

func (s *chatMultiStep) Run(ctx context.Context, baseURL string, modelID string) *api.ScenarioResult {
	url := fmt.Sprintf("%s/v1/chat/completions", baseURL)
	client := http.DefaultClient
	temperature := 0.7
	result := api.NewResult()

	// Step 1: Send the full document plus a simple question about it.
	messages := []api.ChatMessage{
		{Role: api.RoleUser, Content: api.TextContent(
			brindlemarkGuide + "\n\nUsing only the travel guide above, what is the capital city of Brindlemark? Answer concisely.")},
	}

	resp1, err := doChatRequest(ctx, client, url, modelID, &temperature, messages)
	if err != nil {
		result.Fail("single turn chat completion", fmt.Sprintf("%v", err))
		return result
	}

	if api.TextFromContent(resp1.Message.Content) == "" {
		result.Fail("single turn chat completion", "response content is empty")
		return result
	}

	result.Pass("single turn chat completion")

	// Append assistant response to conversation history.
	messages = append(messages, resp1.Message)

	// Step 2: Follow-up question — tests context recall.
	messages = append(messages, api.ChatMessage{
		Role:    api.RoleUser,
		Content: api.TextContent("And which river runs through that city? Answer concisely."),
	})

	resp2, err := doChatRequest(ctx, client, url, modelID, &temperature, messages)
	if err != nil {
		result.Fail("multi-turn context recall", fmt.Sprintf("%v", err))
		return result
	}

	content2 := api.TextFromContent(resp2.Message.Content)
	if content2 == "" {
		result.Fail("multi-turn context recall", "response content is empty")
		return result
	}

	if !strings.Contains(strings.ToLower(content2), "quillsong") {
		result.Fail("multi-turn context recall", fmt.Sprintf("response did not contain 'Quillsong'. Response: %s", content2))
		return result
	}

	result.Pass("multi-turn context recall")

	// The repeated document prefix across the two turns should produce a cache read.
	if resp2.Usage.CacheReadTokens > 0 {
		result.Pass(fmt.Sprintf("prompt cache read observed (%d tokens)", resp2.Usage.CacheReadTokens))
	} else {
		result.Fail("prompt cache read", "no cache read observed — verify the provider/model supports prompt caching")
	}

	return result
}

// doChatRequest sends a chat completion request and returns the decoded response.
func doChatRequest(ctx context.Context, client *http.Client, url, modelID string, temperature *float64, messages []api.ChatMessage) (*api.ChatResponse, error) {
	reqBody := api.ChatRequest{
		Model:       modelID,
		Temperature: temperature,
		Messages:    messages,
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
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var chatResp api.ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &chatResp, nil
}
