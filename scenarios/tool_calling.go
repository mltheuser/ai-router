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
	Register(&toolCalling{})
}

type toolCalling struct{}

func (s *toolCalling) Name() string {
	return "tool_calling"
}

func (s *toolCalling) Description() string {
	return "Verifies tool calling (single and parallel): model invokes tools and incorporates the results"
}

func (s *toolCalling) RequiredCapabilities() []api.Capability {
	return []api.Capability{api.CapabilityChat, api.CapabilityTools}
}

func (s *toolCalling) Run(ctx context.Context, baseURL string, modelID string) *api.ScenarioResult {
	url := fmt.Sprintf("%s/v1/chat/completions", baseURL)
	client := http.DefaultClient
	result := api.NewResult()

	// Define "add" and "multiply" tools.
	tools := []api.ToolDefinition{
		{
			Name:        "add",
			Description: "Add two integers and return the sum",
			Parameters: map[string]interface{}{
				"type":     "object",
				"required": []string{"a", "b"},
				"properties": map[string]interface{}{
					"a": map[string]interface{}{"type": "integer", "description": "First number"},
					"b": map[string]interface{}{"type": "integer", "description": "Second number"},
				},
			},
		},
		{
			Name:        "multiply",
			Description: "Multiply two integers and return the product",
			Parameters: map[string]interface{}{
				"type":     "object",
				"required": []string{"a", "b"},
				"properties": map[string]interface{}{
					"a": map[string]interface{}{"type": "integer", "description": "First number"},
					"b": map[string]interface{}{"type": "integer", "description": "Second number"},
				},
			},
		},
	}

	// Expected tool results.
	toolResults := map[string]string{
		"add":      "5",
		"multiply": "6",
	}

	// Step 1: Ask the model to compute both. A parallel-capable model calls both tools at once.
	messages := []api.ChatMessage{
		{Role: api.RoleUser, Content: api.TextContent("What is 2 + 3 and 2 * 3? You must use the add tool and the multiply tool to compute this.")},
	}

	resp, err := doToolRequest(ctx, client, url, api.ChatRequest{
		Model:    modelID,
		Messages: messages,
		Tools:    tools,
	})
	if err != nil {
		result.Fail("tool invocation request", fmt.Sprintf("%v", err))
		return result
	}

	// Verify the model requested at least one tool call.
	if resp.FinishReason != api.FinishReasonToolCalls {
		result.Fail("tool invocation request", fmt.Sprintf("expected finish_reason 'tool_calls', got '%s'", resp.FinishReason))
		return result
	}
	if len(resp.Message.ToolCalls) == 0 {
		result.Fail("tool invocation request", "no tool_calls in response")
		return result
	}

	// Classify tool calls from the first response.
	firstCalls := resp.Message.ToolCalls
	hasAdd, hasMultiply := findToolCalls(firstCalls)

	// Check: parallel tool calling — both tools called in a single response.
	if hasAdd && hasMultiply {
		result.Pass("parallel tool calling")
	} else {
		missing := "multiply"
		if !hasAdd {
			missing = "add"
		}
		result.Fail("parallel tool calling", fmt.Sprintf("only got one tool call, missing '%s'", missing))
	}

	// Check: tool invocation request — at least one valid tool call.
	if !hasAdd && !hasMultiply {
		result.Fail("tool invocation request", fmt.Sprintf("no recognized tool calls; got: %v", toolCallNames(firstCalls)))
		return result
	}
	result.Pass("tool invocation request")

	// Step 2: Feed tool results back, looping up to 2 iterations for sequential models.
	const maxIterations = 2
	for i := 0; i < maxIterations; i++ {
		// Append the assistant message with its tool calls.
		messages = append(messages, resp.Message)

		// Build tool result messages for every recognized call.
		for _, tc := range resp.Message.ToolCalls {
			if res, ok := toolResults[tc.Function.Name]; ok {
				messages = append(messages, api.ChatMessage{
					Role:       api.RoleTool,
					Content:    api.TextContent(res),
					ToolCallID: tc.ID,
				})
			}
		}

		// Follow-up request.
		resp, err = doToolRequest(ctx, client, url, api.ChatRequest{
			Model:    modelID,
			Messages: messages,
			Tools:    tools,
		})
		if err != nil {
			result.Fail("tool result incorporation", fmt.Sprintf("%v", err))
			return result
		}

		// If model is done calling tools, break to final verification.
		if resp.FinishReason != api.FinishReasonToolCalls {
			break
		}
	}

	// Check: tool result incorporation — final answer must contain both results.
	content := api.TextFromContent(resp.Message.Content)
	if content == "" {
		result.Fail("tool result incorporation", "final response content is empty")
		return result
	}

	has5 := strings.Contains(content, "5")
	has6 := strings.Contains(content, "6")
	if !has5 || !has6 {
		var missing []string
		if !has5 {
			missing = append(missing, "'5' (add result)")
		}
		if !has6 {
			missing = append(missing, "'6' (multiply result)")
		}
		result.Fail("tool result incorporation", fmt.Sprintf("final response missing %s. Response: %s", strings.Join(missing, " and "), content))
		return result
	}

	result.Pass("tool result incorporation")

	return result
}

// findToolCalls checks whether the tool calls contain an "add" and/or "multiply" call.
func findToolCalls(calls []api.ToolCall) (hasAdd, hasMultiply bool) {
	for _, tc := range calls {
		switch tc.Function.Name {
		case "add":
			hasAdd = true
		case "multiply":
			hasMultiply = true
		}
	}
	return
}

// toolCallNames returns the function names from a slice of tool calls (for error messages).
func toolCallNames(calls []api.ToolCall) []string {
	names := make([]string, len(calls))
	for i, tc := range calls {
		names[i] = tc.Function.Name
	}
	return names
}

// doToolRequest sends a chat completion request and returns the decoded response.
func doToolRequest(ctx context.Context, client *http.Client, url string, reqBody api.ChatRequest) (*api.ChatResponse, error) {
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
