package ollama

import (
	"context"
	"fmt"
	"strconv"

	"github.com/mltheuser/ai-router/router/chat"
)

// --- Ollama wire types (request) ---

type ollamaChatRequest struct {
	Model    string                 `json:"model"`
	Messages []ollamaRequestMessage `json:"messages"`
	Stream   bool                   `json:"stream"`
	Format   interface{}            `json:"format,omitempty"`
	Options  *ollamaOptions         `json:"options,omitempty"`
	Think    interface{}            `json:"think"`
	Tools    []ollamaToolDefinition `json:"tools,omitempty"`
}

type ollamaToolDefinition struct {
	Type     string              `json:"type"`
	Function chat.ToolDefinition `json:"function"`
}

// ollamaRequestMessage names the tool a tool result answers by "tool_name",
// not "tool_call_id".
type ollamaRequestMessage struct {
	Role      string                  `json:"role"`
	Content   string                  `json:"content"`
	Images    []string                `json:"images,omitempty"`
	ToolCalls []ollamaRequestToolCall `json:"tool_calls,omitempty"`
	ToolName  string                  `json:"tool_name,omitempty"`
}

// ollamaRequestToolCall identifies a call by an "index" inside function, not
// by an "id" at the top level.
type ollamaRequestToolCall struct {
	Type     string                    `json:"type"`
	Function ollamaRequestToolCallFunc `json:"function"`
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
	DoneReason      string        `json:"done_reason"`
	PromptEvalCount int           `json:"prompt_eval_count"`
	EvalCount       int           `json:"eval_count"`
	TotalDuration   int64         `json:"total_duration"`
	LoadDuration    int64         `json:"load_duration"`
	EvalDuration    int64         `json:"eval_duration"`
}

type ollamaMessage struct {
	Role      string                   `json:"role"`
	Content   string                   `json:"content"`
	Thinking  string                   `json:"thinking,omitempty"`
	ToolCalls []ollamaResponseToolCall `json:"tool_calls,omitempty"`
}

type ollamaResponseToolCall struct {
	Type     string                     `json:"type"`
	Function ollamaResponseToolCallFunc `json:"function"`
}

type ollamaResponseToolCallFunc struct {
	Index     int                    `json:"index"`
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// --- Chat implementation ---

func (p *Provider) Chat(ctx context.Context, req *chat.Request) (*chat.Response, error) {
	ollamaReq := ollamaChatRequest{
		Model:    req.Model,
		Messages: toOllamaMessages(req.Messages),
		Stream:   false,
		Tools:    wrapTools(req.Tools),
	}

	if req.ReasoningEffort == chat.ReasoningEffortNone {
		ollamaReq.Think = false
	} else {
		ollamaReq.Think = string(req.ReasoningEffort)
	}

	if req.Temperature != nil || req.TopP != nil || req.MaxTokens != nil {
		ollamaReq.Options = &ollamaOptions{
			Temperature: req.Temperature,
			TopP:        req.TopP,
			NumPredict:  req.MaxTokens,
		}
	}

	if req.ResponseFormat != nil {
		if req.ResponseFormat.Type == chat.ResponseFormatJSONSchema && req.ResponseFormat.JSONSchema != nil {
			ollamaReq.Format = req.ResponseFormat.JSONSchema.Schema
		} else if req.ResponseFormat.Type == "json_object" {
			ollamaReq.Format = "json"
		}
	}

	var ollamaResp ollamaChatResponse

	if err := p.client.Post(ctx, "/api/chat", ollamaReq, &ollamaResp); err != nil {
		return nil, err
	}

	return mapResponse(&ollamaResp), nil
}

// --- Request translation ---

func wrapTools(tools []chat.ToolDefinition) []ollamaToolDefinition {
	if len(tools) == 0 {
		return nil
	}
	result := make([]ollamaToolDefinition, len(tools))
	for i, t := range tools {
		result[i] = ollamaToolDefinition{Type: "function", Function: t}
	}
	return result
}

func toOllamaMessages(messages []chat.Message) []ollamaRequestMessage {
	result := make([]ollamaRequestMessage, len(messages))
	for i, m := range messages {
		om := ollamaRequestMessage{
			Role:    string(m.Role),
			Content: chat.TextFromContent(m.Content),
			Images:  chat.ImagesFromContent(m.Content),
		}

		switch m.Role {
		case chat.RoleAssistant:
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
		case chat.RoleTool:
			if m.ToolCallID != "" {
				om.ToolName = findToolName(messages, m.ToolCallID)
			}
		}

		result[i] = om
	}
	return result
}

// --- Response translation ---

func mapResponse(ollamaResp *ollamaChatResponse) *chat.Response {
	resp := chat.Response{
		Model: ollamaResp.Model,
		Usage: chat.Usage{
			PromptTokens:     ollamaResp.PromptEvalCount,
			CompletionTokens: ollamaResp.EvalCount,
			TotalTokens:      ollamaResp.PromptEvalCount + ollamaResp.EvalCount,
			// Ollama reports no cached tokens.
		},
		FinishReason: mapFinishReason(ollamaResp.DoneReason),
		Message: chat.Message{
			Role:    chat.Role(ollamaResp.Message.Role),
			Content: chat.TextContent(ollamaResp.Message.Content),
		},
	}

	if ollamaResp.Message.Thinking != "" {
		resp.Message.ReasoningContent = ollamaResp.Message.Thinking
	}

	// Ollama identifies tool calls by index, which becomes the ID.
	if len(ollamaResp.Message.ToolCalls) > 0 {
		for _, tc := range ollamaResp.Message.ToolCalls {
			resp.Message.ToolCalls = append(resp.Message.ToolCalls, chat.ToolCall{
				ID: fmt.Sprintf("%d", tc.Function.Index),
				Function: chat.ToolCallFunction{
					Name:      tc.Function.Name,
					Arguments: tc.Function.Arguments,
				},
			})
		}
		resp.FinishReason = chat.FinishReasonToolCalls
	}

	return &resp
}

// --- Helpers ---

// mapFinishReason maps Ollama's done_reason to the shared FinishReason.
// "length" indicates truncation; "stop", lifecycle values (e.g. "load",
// "unload"), and empty all resolve to stop.
func mapFinishReason(doneReason string) chat.FinishReason {
	switch doneReason {
	case "length":
		return chat.FinishReasonLength
	default:
		return chat.FinishReasonStop
	}
}

// findToolName searches backwards through messages for the tool call matching
// the given ID and returns its function name.
func findToolName(messages []chat.Message, toolCallID string) string {
	for j := len(messages) - 1; j >= 0; j-- {
		for _, tc := range messages[j].ToolCalls {
			if tc.ID == toolCallID {
				return tc.Function.Name
			}
		}
	}
	return ""
}
