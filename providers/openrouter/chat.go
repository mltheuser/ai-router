package openrouter

import (
	"context"
	"encoding/json"

	"github.com/mltheuser/ai-router/api"
)

// --- OpenRouter wire types (request) ---

type openRouterChatRequest struct {
	Model            string                      `json:"model"`
	Messages         []openRouterRequestMessage  `json:"messages"`
	FrequencyPenalty *float64                    `json:"frequency_penalty,omitempty"`
	MaxTokens        *int                        `json:"max_tokens,omitempty"`
	PresencePenalty  *float64                    `json:"presence_penalty,omitempty"`
	Temperature      *float64                    `json:"temperature,omitempty"`
	TopP             *float64                    `json:"top_p,omitempty"`
	ResponseFormat   *api.ResponseFormat         `json:"response_format,omitempty"`
	ReasoningEffort  *string                     `json:"reasoning_effort,omitempty"`
	Tools            []openRouterToolDefinition  `json:"tools,omitempty"`
}

// openRouterToolDefinition wraps our flat ToolDefinition in OpenRouter's {"type":"function","function":{...}} format.
type openRouterToolDefinition struct {
	Type     string             `json:"type"`
	Function api.ToolDefinition `json:"function"`
}

// openRouterRequestMessage is the outgoing message format for OpenRouter.
type openRouterRequestMessage struct {
	Role       string                      `json:"role"`
	Content    string                      `json:"content"`
	ToolCalls  []openRouterRequestToolCall `json:"tool_calls,omitempty"`
	ToolCallID string                      `json:"tool_call_id,omitempty"`
}

// openRouterRequestToolCall is the outgoing tool call format for OpenRouter.
// OpenRouter (OpenAI-compatible) serializes arguments as a JSON string.
type openRouterRequestToolCall struct {
	ID       string                         `json:"id"`
	Type     string                         `json:"type"`
	Function openRouterRequestToolCallFunc  `json:"function"`
}

type openRouterRequestToolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// --- OpenRouter wire types (response) ---

type openRouterChatChoice struct {
	Index        int                    `json:"index"`
	Message      openRouterMessage      `json:"message"`
	FinishReason string                 `json:"finish_reason"`
}

type openRouterMessage struct {
	Role      string                    `json:"role"`
	Content   string                    `json:"content"`
	Reasoning string                    `json:"reasoning,omitempty"`
	Thinking  string                    `json:"thinking,omitempty"`
	ToolCalls []openRouterResponseToolCall `json:"tool_calls,omitempty"`
}

type openRouterResponseToolCall struct {
	ID       string                          `json:"id"`
	Type     string                          `json:"type"`
	Function openRouterResponseToolCallFunc  `json:"function"`
}

type openRouterResponseToolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openRouterUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
	CompletionTokensDetails *struct {
		ReasoningTokens int `json:"reasoning_tokens"`
	} `json:"completion_tokens_details,omitempty"`
}

type openRouterChatResponse struct {
	ID      string                 `json:"id"`
	Object  string                 `json:"object"`
	Created int64                  `json:"created"`
	Model   string                 `json:"model"`
	Choices []openRouterChatChoice `json:"choices"`
	Usage   openRouterUsage        `json:"usage"`
}

// --- Chat implementation ---

func (p *Provider) Chat(ctx context.Context, req *api.ChatRequest) (*api.ChatResponse, error) {
	orReq := toOpenRouterRequest(req)

	var orResp openRouterChatResponse
	if err := p.client.post(ctx, "/chat/completions", orReq, &orResp); err != nil {
		return nil, err
	}

	return mapOpenRouterResponse(&orResp), nil
}

// --- Request translation ---

func toOpenRouterRequest(req *api.ChatRequest) *openRouterChatRequest {
	orReq := &openRouterChatRequest{
		Model:            req.Model,
		FrequencyPenalty: req.FrequencyPenalty,
		MaxTokens:        req.MaxTokens,
		PresencePenalty:  req.PresencePenalty,
		Temperature:      req.Temperature,
		TopP:             req.TopP,
		ResponseFormat:   req.ResponseFormat,
		ReasoningEffort:  req.ReasoningEffort,
	}

	// Wrap tools in {"type":"function","function":{...}}
	for _, t := range req.Tools {
		orReq.Tools = append(orReq.Tools, openRouterToolDefinition{
			Type:     "function",
			Function: t,
		})
	}

	// Transform messages
	for _, m := range req.Messages {
		msg := openRouterRequestMessage{
			Role:    m.Role,
			Content: m.Content,
		}

		switch m.Role {
		case api.RoleAssistant:
			// Convert tool calls: serialize arguments map → JSON string
			for _, tc := range m.ToolCalls {
				argsJSON, _ := json.Marshal(tc.Function.Arguments)
				msg.ToolCalls = append(msg.ToolCalls, openRouterRequestToolCall{
					ID:   tc.ID,
					Type: "function",
					Function: openRouterRequestToolCallFunc{
						Name:      tc.Function.Name,
						Arguments: string(argsJSON),
					},
				})
			}
		case api.RoleTool:
			// OpenRouter uses "tool_call_id" to match results to calls.
			msg.ToolCallID = m.ToolCallID
		}

		orReq.Messages = append(orReq.Messages, msg)
	}

	return orReq
}

// --- Response translation ---

func mapOpenRouterResponse(orResp *openRouterChatResponse) *api.ChatResponse {
	resp := api.ChatResponse{
		Model: orResp.Model,
		Usage: api.ChatUsage{
			PromptTokens:     orResp.Usage.PromptTokens,
			CompletionTokens: orResp.Usage.CompletionTokens,
			TotalTokens:      orResp.Usage.TotalTokens,
		},
	}

	if orResp.Usage.CompletionTokensDetails != nil {
		resp.Usage.ReasoningTokens = orResp.Usage.CompletionTokensDetails.ReasoningTokens
	}

	if len(orResp.Choices) > 0 {
		c := orResp.Choices[0]

		reasoning := c.Message.Reasoning
		if reasoning == "" {
			reasoning = c.Message.Thinking
		}

		resp.Choice = api.ChatChoice{
			Message: api.ChatMessage{
				Role:             c.Message.Role,
				Content:          c.Message.Content,
				ReasoningContent: reasoning,
			},
			FinishReason: c.FinishReason,
		}

		// Map tool calls: parse JSON-string arguments → map
		if len(c.Message.ToolCalls) > 0 {
			resp.Choice.FinishReason = "tool_calls"
			for _, tc := range c.Message.ToolCalls {
				resp.Choice.Message.ToolCalls = append(resp.Choice.Message.ToolCalls, api.ToolCall{
					ID: tc.ID,
					Function: api.ToolCallFunction{
						Name:      tc.Function.Name,
						Arguments: parseArguments(tc.Function.Arguments),
					},
				})
			}
		}
	}

	return &resp
}

// parseArguments converts a JSON-encoded arguments string to a map.
func parseArguments(raw string) map[string]interface{} {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &args); err != nil {
		return map[string]interface{}{"raw": raw}
	}
	return args
}
