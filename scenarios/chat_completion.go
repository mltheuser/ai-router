package scenarios

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mltheuser/ai-router/api"
)

func init() {
	Register(&ChatMultiStep{})
}

type ChatMultiStep struct{}

func (s *ChatMultiStep) Name() string {
	return "chat_multi_step"
}

func (s *ChatMultiStep) Description() string {
	return "Verifies basic chat completion and multi-step conversation recall"
}

func (s *ChatMultiStep) RequiredCapabilities() []api.Capability {
	return []api.Capability{api.CapabilityChat}
}

func (s *ChatMultiStep) Run(ctx context.Context, baseURL string, modelID string) *api.ScenarioResult {
	url := fmt.Sprintf("%s/v1/chat/completions", baseURL)
	client := http.DefaultClient
	temperature := 0.7
	result := api.NewResult()

	// Step 1: Initial message — establishes a fact for later recall.
	messages := []api.ChatMessage{
		{Role: "user", Content: api.TextContent("My favorite color is blue. Remember this.")},
	}

	resp1, err := doChatRequest(ctx, client, url, modelID, &temperature, messages)
	if err != nil {
		result.Fail("single turn chat completion", fmt.Sprintf("%v", err))
		return result
	}

	if api.TextFromContent(resp1.Choice.Message.Content) == "" {
		result.Fail("single turn chat completion", "response content is empty")
		return result
	}

	result.Pass("single turn chat completion")

	// Append assistant response to conversation history.
	messages = append(messages, resp1.Choice.Message)

	// Step 2: Follow-up question — tests context recall.
	messages = append(messages, api.ChatMessage{
		Role:    "user",
		Content: api.TextContent("What is my favorite color?"),
	})

	resp2, err := doChatRequest(ctx, client, url, modelID, &temperature, messages)
	if err != nil {
		result.Fail("multi-turn context recall", fmt.Sprintf("%v", err))
		return result
	}

	content2 := api.TextFromContent(resp2.Choice.Message.Content)
	if content2 == "" {
		result.Fail("multi-turn context recall", "response content is empty")
		return result
	}

	if !strings.Contains(strings.ToLower(content2), "blue") {
		result.Fail("multi-turn context recall", fmt.Sprintf("response did not contain 'blue'. Response: %s", content2))
		return result
	}

	result.Pass("multi-turn context recall")

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
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var chatResp api.ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %v", err)
	}

	return &chatResp, nil
}
