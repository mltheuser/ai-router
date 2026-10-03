package openrouter

import (
	"context"
	"encoding/json"

	"github.com/mltheuser/ai-router/router/chat"
)

// --- OpenRouter wire types (request) ---

type openRouterChatRequest struct {
	Model            string                     `json:"model"`
	Messages         []openRouterRequestMessage `json:"messages"`
	FrequencyPenalty *float64                   `json:"frequency_penalty,omitempty"`
	MaxTokens        *int                       `json:"max_tokens,omitempty"`
	PresencePenalty  *float64                   `json:"presence_penalty,omitempty"`
	Temperature      *float64                   `json:"temperature,omitempty"`
	TopP             *float64                   `json:"top_p,omitempty"`
	ResponseFormat   *openRouterResponseFormat  `json:"response_format,omitempty"`
	ReasoningEffort  string                     `json:"reasoning_effort"`
	Tools            []openRouterToolDefinition `json:"tools,omitempty"`
}

type openRouterResponseFormat struct {
	Type       chat.ResponseFormatType `json:"type"`
	JSONSchema *openRouterJSONSchema   `json:"json_schema"`
}

type openRouterJSONSchema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
	Strict      bool                   `json:"strict,omitempty"`
}

type openRouterToolDefinition struct {
	Type     string              `json:"type"`
	Function chat.ToolDefinition `json:"function"`
}

type openRouterRequestMessage struct {
	Role       string                      `json:"role"`
	Content    []openRouterContentPart     `json:"content"`
	ToolCalls  []openRouterRequestToolCall `json:"tool_calls,omitempty"`
	ToolCallID string                      `json:"tool_call_id,omitempty"`
}

// openRouterContentPart has exactly one of Text or ImageURL set, matching Type.
type openRouterContentPart struct {
	Type         string                     `json:"type"` // "text" or "image_url"
	Text         string                     `json:"text,omitempty"`
	ImageURL     *openRouterContentImageURL `json:"image_url,omitempty"`
	CacheControl *openRouterCacheControl    `json:"cache_control,omitempty"`
}

type openRouterCacheControl struct {
	Type string `json:"type"`
}

type openRouterContentImageURL struct {
	URL string `json:"url"`
}

// openRouterRequestToolCall carries the arguments as a JSON string.
type openRouterRequestToolCall struct {
	ID       string                        `json:"id"`
	Type     string                        `json:"type"`
	Function openRouterRequestToolCallFunc `json:"function"`
}

type openRouterRequestToolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// --- OpenRouter wire types (response) ---

type openRouterChatChoice struct {
	Index        int               `json:"index"`
	Message      openRouterMessage `json:"message"`
	FinishReason string            `json:"finish_reason"`
}

type openRouterMessage struct {
	Role      string                       `json:"role"`
	Content   string                       `json:"content"`
	Reasoning string                       `json:"reasoning,omitempty"`
	Thinking  string                       `json:"thinking,omitempty"`
	ToolCalls []openRouterResponseToolCall `json:"tool_calls,omitempty"`
}

type openRouterResponseToolCall struct {
	ID       string                         `json:"id"`
	Type     string                         `json:"type"`
	Function openRouterResponseToolCallFunc `json:"function"`
}

type openRouterResponseToolCallFunc struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type openRouterUsage struct {
	PromptTokens            int `json:"prompt_tokens"`
	CompletionTokens        int `json:"completion_tokens"`
	TotalTokens             int `json:"total_tokens"`
	CompletionTokensDetails *struct {
		ReasoningTokens int `json:"reasoning_tokens"`
	} `json:"completion_tokens_details,omitempty"`
	PromptTokensDetails *struct {
		CachedTokens int `json:"cached_tokens"`
	} `json:"prompt_tokens_details,omitempty"`
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

func (p *Provider) Chat(ctx context.Context, req *chat.Request) (*chat.Response, error) {
	orReq := toOpenRouterRequest(req)

	var orResp openRouterChatResponse
	if err := p.client.Post(ctx, "/chat/completions", orReq, &orResp); err != nil {
		return nil, err
	}

	return mapOpenRouterResponse(&orResp), nil
}

// --- Request translation ---

func toOpenRouterRequest(req *chat.Request) *openRouterChatRequest {
	orReq := &openRouterChatRequest{
		Model:            req.Model,
		FrequencyPenalty: req.FrequencyPenalty,
		MaxTokens:        req.MaxTokens,
		PresencePenalty:  req.PresencePenalty,
		Temperature:      req.Temperature,
		TopP:             req.TopP,
		ResponseFormat:   toOpenRouterResponseFormat(req.ResponseFormat),
		ReasoningEffort:  string(req.ReasoningEffort),
	}

	for _, t := range req.Tools {
		orReq.Tools = append(orReq.Tools, openRouterToolDefinition{
			Type:     "function",
			Function: t,
		})
	}

	for _, m := range req.Messages {
		msg := openRouterRequestMessage{
			Role:    string(m.Role),
			Content: toOpenRouterContent(m.Content),
		}

		switch m.Role {
		case chat.RoleAssistant:
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
		case chat.RoleTool:
			msg.ToolCallID = m.ToolCallID
		}

		orReq.Messages = append(orReq.Messages, msg)
	}

	// Always request prompt caching: it is transparent to clients. One
	// breakpoint on the last content block caches the whole prompt up to it;
	// models that cache implicitly ignore it, and prompts below a model's
	// minimum cacheable length are not cached. The walk goes backwards past
	// messages without content parts, such as bare tool calls.
	for i := len(orReq.Messages) - 1; i >= 0; i-- {
		if parts := orReq.Messages[i].Content; len(parts) > 0 {
			parts[len(parts)-1].CacheControl = &openRouterCacheControl{Type: "ephemeral"}
			break
		}
	}

	return orReq
}

func toOpenRouterContent(parts []chat.ContentPart) []openRouterContentPart {
	result := make([]openRouterContentPart, 0, len(parts))
	for _, p := range parts {
		switch p.Type {
		case chat.ContentPartText:
			result = append(result, openRouterContentPart{
				Type: "text",
				Text: p.Text,
			})
		case chat.ContentPartImage:
			result = append(result, openRouterContentPart{
				Type: "image_url",
				ImageURL: &openRouterContentImageURL{
					URL: "data:" + p.MimeType + ";base64," + p.Base64Data,
				},
			})
		}
	}
	return result
}

func toOpenRouterResponseFormat(rf *chat.ResponseFormat) *openRouterResponseFormat {
	if rf == nil {
		return nil
	}

	orf := &openRouterResponseFormat{
		Type: rf.Type,
	}

	if rf.JSONSchema != nil {
		orf.JSONSchema = &openRouterJSONSchema{
			Name:        rf.JSONSchema.Name,
			Description: rf.JSONSchema.Description,
			Schema:      rf.JSONSchema.Schema,
			Strict:      true,
		}
	}

	return orf
}

// --- Response translation ---

func mapOpenRouterResponse(orResp *openRouterChatResponse) *chat.Response {
	resp := chat.Response{
		Model: orResp.Model,
		Usage: chat.Usage{
			PromptTokens:     orResp.Usage.PromptTokens,
			CompletionTokens: orResp.Usage.CompletionTokens,
			TotalTokens:      orResp.Usage.TotalTokens,
		},
	}

	if orResp.Usage.CompletionTokensDetails != nil {
		resp.Usage.ReasoningTokens = orResp.Usage.CompletionTokensDetails.ReasoningTokens
	}

	// prompt_tokens already includes cached tokens.
	if orResp.Usage.PromptTokensDetails != nil {
		resp.Usage.CacheReadTokens = orResp.Usage.PromptTokensDetails.CachedTokens
	}

	if len(orResp.Choices) > 0 {
		c := orResp.Choices[0]

		reasoning := c.Message.Reasoning
		if reasoning == "" {
			reasoning = c.Message.Thinking
		}

		resp.Message = chat.Message{
			Role:             chat.Role(c.Message.Role),
			Content:          chat.TextContent(c.Message.Content),
			ReasoningContent: reasoning,
		}
		resp.FinishReason = chat.FinishReason(c.FinishReason)

		if len(c.Message.ToolCalls) > 0 {
			resp.FinishReason = chat.FinishReasonToolCalls
			for _, tc := range c.Message.ToolCalls {
				resp.Message.ToolCalls = append(resp.Message.ToolCalls, chat.ToolCall{
					ID: tc.ID,
					Function: chat.ToolCallFunction{
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
