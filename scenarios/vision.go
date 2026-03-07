package scenarios

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/mltheuser/ai-router/api"
)

func init() {
	testImageBase64 = base64.StdEncoding.EncodeToString(testImageBytes)
	Register(&VisionDescription{})
}

type VisionDescription struct{}

func (s *VisionDescription) Name() string {
	return "vision_description"
}

func (s *VisionDescription) Description() string {
	return "Verifies that vision models can describe an image"
}

func (s *VisionDescription) RequiredCapabilities() []api.Capability {
	return []api.Capability{api.CapabilityChat, api.CapabilityVision}
}

//go:embed resources/apple.png
var testImageBytes []byte
var testImageBase64 string

func (s *VisionDescription) Run(ctx context.Context, baseURL string, modelID string) *api.ScenarioResult {
	url := fmt.Sprintf("%s/v1/chat/completions", baseURL)
	result := api.NewResult()

	reqBody := api.ChatRequest{
		Model: modelID,
		Messages: []api.ChatMessage{
			{
				Role: api.RoleUser,
				Content: []api.ContentPart{
					{Type: api.ContentPartText, Text: "What fruit do you see in the image? Be concise."},
					{Type: api.ContentPartImage, MimeType: "image/png", Base64Data: testImageBase64},
				},
			},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		result.Fail("vision image description", fmt.Sprintf("failed to marshal request: %v", err))
		return result
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		result.Fail("vision image description", fmt.Sprintf("failed to create request: %v", err))
		return result
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		result.Fail("vision image description", fmt.Sprintf("request failed: %v", err))
		return result
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		result.Fail("vision image description", fmt.Sprintf("unexpected status code: %d", resp.StatusCode))
		return result
	}

	var chatResp api.ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		result.Fail("vision image description", fmt.Sprintf("failed to decode response: %v", err))
		return result
	}

	content := api.TextFromContent(chatResp.Choice.Message.Content)
	if content == "" {
		result.Fail("vision image description", "response content is empty")
		return result
	}

	if !strings.Contains(strings.ToLower(content), "apple") {
		result.Fail("vision image description", fmt.Sprintf("response does not contain 'apple': %s", content))
		return result
	}

	result.Pass("vision image description")

	return result
}
