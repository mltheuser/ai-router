package ollama

import (
	"context"
	"fmt"
	"strconv"
	"strings"

	"github.com/mltheuser/ai-router/api"
)

// --- Ollama wire types (request) ---

type ollamaChatRequest struct {
	Model    string                 `json:"model"`
	Messages []ollamaRequestMessage `json:"messages"`
	Stream   bool                   `json:"stream"`
	Format   interface{}            `json:"format,omitempty"`
	Options  *ollamaOptions         `json:"options,omitempty"`
	Think    interface{}            `json:"think,omitempty"`
	Tools    []ollamaToolDefinition `json:"tools,omitempty"`
}

// ollamaToolDefinition wraps our flat ToolDefinition in Ollama's {"type":"function","function":{...}} format.
type ollamaToolDefinition struct {
	Type     string           `json:"type"`
	Function api.ToolDefinition `json:"function"`
}

// ollamaRequestMessage is the outgoing message format for Ollama.
// Ollama uses "tool_name" on tool result messages (not "tool_call_id").
type ollamaRequestMessage struct {
	Role      string                `json:"role"`
	Content   string                `json:"content"`
	ToolCalls []ollamaRequestToolCall `json:"tool_calls,omitempty"`
	ToolName  string                `json:"tool_name,omitempty"`
}

// ollamaRequestToolCall is the outgoing tool call format for Ollama.
// Ollama uses an "index" field inside function, not "id" at the top level.
type ollamaRequestToolCall struct {
	Type     string                     `json:"type"`
	Function ollamaRequestToolCallFunc  `json:"function"`
}

type ollamaRequestToolCallFunc struct {
	Index     int                    `json:"index"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ollamaOptions struct {
	Temperature *float64 `json:"temperature,omitempty"`
	TopP        *float64 `json:"top_p,omitempty"`
	NumPredict  *int     `json:"num_predict,omitempty"`
	Stop        []string `json:"stop,omitempty"`
}

// --- Ollama wire types (response) ---

type ollamaChatResponse struct {
	Model           string        `json:"model"`
	CreatedAt       string        `json:"created_at"`
	Message         ollamaMessage `json:"message"`
	Done            bool          `json:"done"`
	PromptEvalCount int           `json:"prompt_eval_count"`
	EvalCount       int           `json:"eval_count"`
	TotalDuration   int64         `json:"total_duration"`
	LoadDuration    int64         `json:"load_duration"`
	EvalDuration    int64         `json:"eval_duration"`
}

type ollamaMessage struct {
	Role      string              `json:"role"`
	Content   string              `json:"content"`
	Thinking  string              `json:"thinking,omitempty"`
	ToolCalls []ollamaResponseToolCall `json:"tool_calls,omitempty"`
}

type ollamaResponseToolCall struct {
	Type     string                      `json:"type"`
	Function ollamaResponseToolCallFunc  `json:"function"`
}

type ollamaResponseToolCallFunc struct {
	Index     int                    `json:"index"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// --- Chat implementation ---

func (p *Provider) Chat(ctx context.Context, req *api.ChatRequest) (*api.ChatResponse, error) {
	ollamaReq := ollamaChatRequest{
		Model:    req.Model,
		Messages: toOllamaMessages(req.Messages),
		Stream:   false,
		Tools:    wrapTools(req.Tools),
	}

	// Map generic ReasoningEffort to Ollama's "think" parameter.
	if req.ReasoningEffort != nil {
		if *req.ReasoningEffort == api.ReasoningEffortNone {
			ollamaReq.Think = false
		} else {
			ollamaReq.Think = *req.ReasoningEffort
		}
	}

	// Map options
	if req.Temperature != nil || req.TopP != nil || req.MaxTokens != nil {
		ollamaReq.Options = &ollamaOptions{
			Temperature: req.Temperature,
			TopP:        req.TopP,
			NumPredict:  req.MaxTokens,
		}
	}

	// Handle Structured Output
	if req.ResponseFormat != nil {
		if req.ResponseFormat.Type == api.ResponseFormatJSONSchema && req.ResponseFormat.JSONSchema != nil {
			ollamaReq.Format = req.ResponseFormat.JSONSchema.Schema
		} else if req.ResponseFormat.Type == "json_object" {
			ollamaReq.Format = "json"
		}
	}

	var ollamaResp ollamaChatResponse

	err := p.client.post(ctx, "/api/chat", ollamaReq, &ollamaResp)
	if err != nil {
		if isUnsupportedThinkValueError(err) && ollamaReq.Think != false {
			ollamaReq.Think = true
			if retryErr := p.client.post(ctx, "/api/chat", ollamaReq, &ollamaResp); retryErr != nil {
				return nil, retryErr
			}
		} else {
			return nil, err
		}
	}

	return mapResponse(&ollamaResp), nil
}

// --- Request translation ---

// wrapTools converts flat ToolDefinitions to Ollama's nested wire format.
func wrapTools(tools []api.ToolDefinition) []ollamaToolDefinition {
	if len(tools) == 0 {
		return nil
	}
	result := make([]ollamaToolDefinition, len(tools))
	for i, t := range tools {
		result[i] = ollamaToolDefinition{Type: "function", Function: t}
	}
	return result
}

// toOllamaMessages transforms shared API messages to Ollama's native format.
func toOllamaMessages(messages []api.ChatMessage) []ollamaRequestMessage {
	result := make([]ollamaRequestMessage, len(messages))
	for i, m := range messages {
		om := ollamaRequestMessage{
			Role:    m.Role,
			Content: m.Content,
		}

		switch m.Role {
		case api.RoleAssistant:
			// Convert tool calls: map string ID → integer index
			for _, tc := range m.ToolCalls {
				idx, _ := strconv.Atoi(tc.ID) // synthetic IDs are stringified indices
				om.ToolCalls = append(om.ToolCalls, ollamaRequestToolCall{
					Type: "function",
					Function: ollamaRequestToolCallFunc{
						Index:     idx,
						Name:      tc.Function.Name,
						Arguments: tc.Function.Arguments,
					},
				})
			}
		case api.RoleTool:
			// Ollama uses "tool_name" to identify which tool the result is for.
			// Look up the function name from the assistant's tool calls using ToolCallID.
			if m.ToolCallID != "" {
				om.ToolName = findToolName(messages, m.ToolCallID)
			}
		}

		result[i] = om
	}
	return result
}

// --- Response translation ---

func mapResponse(ollamaResp *ollamaChatResponse) *api.ChatResponse {
	resp := api.ChatResponse{
		Model: ollamaResp.Model,
		Usage: api.ChatUsage{
			PromptTokens:     ollamaResp.PromptEvalCount,
			CompletionTokens: ollamaResp.EvalCount,
			TotalTokens:      ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
		},
		Choice: api.ChatChoice{
			FinishReason: "stop",
			Message: api.ChatMessage{
				Role:    ollamaResp.Message.Role,
				Content: ollamaResp.Message.Content,
			},
		},
	}

	if ollamaResp.Message.Thinking != "" {
		resp.Choice.Message.ReasoningContent = ollamaResp.Message.Thinking
	}

	// Map tool calls: convert Ollama's index-based calls to ID-based shared format.
	if len(ollamaResp.Message.ToolCalls) > 0 {
		for _, tc := range ollamaResp.Message.ToolCalls {
			resp.Choice.Message.ToolCalls = append(resp.Choice.Message.ToolCalls, api.ToolCall{
				ID: fmt.Sprintf("%d", tc.Function.Index),
				Function: api.ToolCallFunction{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}
		resp.Choice.FinishReason = "tool_calls"
	}

	return &resp
}

// --- Helpers ---

func isUnsupportedThinkValueError(err error) bool {
	msg := err.Error()
	return strings.Contains(msg, "think value") && strings.Contains(msg, "not supported")
}

// findToolName searches backwards through messages for the tool call matching
// the given ID and returns its function name.
func findToolName(messages []api.ChatMessage, toolCallID string) string {
	for j := len(messages) - 1; j >= 0; j-- {
		for _, tc := range messages[j].ToolCalls {
			if tc.ID == toolCallID {
				return tc.Function.Name
			}
		}
	}
	return ""
}
