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
	Register(&StructuredOutput{})
}

type StructuredOutput struct{}

func (s *StructuredOutput) Name() string {
	return "structured_output"
}

func (s *StructuredOutput) Description() string {
	return "Verifies structured output functionality"
}

func (s *StructuredOutput) RequiredCapabilities() []api.Capability {
	return []api.Capability{api.CapabilityChat, api.CapabilityStructuredOutput}
}

func (s *StructuredOutput) Run(ctx context.Context, baseURL string, modelID string) *api.ScenarioResult {
	url := fmt.Sprintf("%s/v1/chat/completions", baseURL)
	result := api.NewResult()

	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location": map[string]interface{}{
				"type": "string",
			},
			"temperature": map[string]interface{}{
				"type": "number",
			},
			"unit": map[string]interface{}{
				"type": "string",
				"enum": []string{"celsius", "fahrenheit"},
			},
		},
		"required":             []string{"location", "temperature", "unit"},
		"additionalProperties": false,
	}

	reqBody := api.ChatRequest{
		Model: modelID,
		Messages: []api.ChatMessage{
			{Role: "user", Content: "It is 25 degrees celsius in Paris."},
		},
		ResponseFormat: &api.ResponseFormat{
			Type: api.ResponseFormatJSONSchema,
			JSONSchema: &api.JSONSchema{
				Name:   "weather_response",
				Schema: schema,
				Strict: true,
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		result.Fail("structured JSON output", fmt.Sprintf("failed to marshal request: %v", err))
		return result
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		result.Fail("structured JSON output", fmt.Sprintf("failed to create request: %v", err))
		return result
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		result.Fail("structured JSON output", fmt.Sprintf("request failed: %v", err))
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		result.Fail("structured JSON output", fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
		return result
	}

	var chatResp api.ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		result.Fail("structured JSON output", fmt.Sprintf("failed to decode response: %v", err))
		return result
	}

	content := chatResp.Choice.Message.Content
	if content == "" {
		result.Fail("structured JSON output", "response content is empty")
		return result
	}

	// Verify that the content is valid JSON matching our structure
	var weather struct {
		Location    string  `json:"location"`
		Temperature float64 `json:"temperature"`
		Unit        string  `json:"unit"`
	}

	if err := json.Unmarshal([]byte(content), &weather); err != nil {
		result.Fail("structured JSON output", fmt.Sprintf("failed to unmarshal structured output: %v. Content: %s", err, content))
		return result
	}

	if weather.Location == "" || weather.Unit == "" {
		result.Fail("structured JSON output", fmt.Sprintf("structured output missing required fields. Content: %s", content))
		return result
	}

	result.Pass("structured JSON output")

	return result
}
