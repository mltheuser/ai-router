package scenarios

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/mltheuser/ai-router/api"
)

func init() {
	Register(&thinkingScenario{})
}

type thinkingScenario struct{}

func (s *thinkingScenario) Name() string {
	return "thinking"
}

func (s *thinkingScenario) Description() string {
	return "Verifies that reasoning models return a reasoning trace"
}

func (s *thinkingScenario) RequiredCapabilities() []api.Capability {
	return []api.Capability{api.CapabilityReasoning}
}

func (s *thinkingScenario) Run(ctx context.Context, baseURL string, modelID string) *api.ScenarioResult {
	url := fmt.Sprintf("%s/v1/chat/completions", baseURL)
	result := api.NewResult()

	// We set reasoning effort to ensure the provider handles the parameter
	effort := "low"
	reqBody := api.ChatRequest{
		Model: modelID,
		Messages: []api.ChatMessage{
			{Role: "user", Content: api.TextContent("Why is the sky blue?")},
		},
		ReasoningEffort: &effort,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		result.Fail("reasoning trace present", fmt.Sprintf("failed to marshal request: %v", err))
		return result
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		result.Fail("reasoning trace present", fmt.Sprintf("failed to create request: %v", err))
		return result
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		result.Fail("reasoning trace present", fmt.Sprintf("request failed: %v", err))
		return result
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		result.Fail("reasoning trace present", fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
		return result
	}

	var chatResp api.ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		result.Fail("reasoning trace present", fmt.Sprintf("failed to decode response: %v", err))
		return result
	}

	msg := chatResp.Choice.Message
	if msg.ReasoningContent == "" {
		result.Fail("reasoning trace present", fmt.Sprintf("expected reasoning content, got empty string. Content was: %s", api.TextFromContent(msg.Content)))
		return result
	}

	result.Pass("reasoning trace present")

	return result
}
