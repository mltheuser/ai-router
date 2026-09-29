package anthropic

import (
	"context"
	"encoding/json"
	"strings"

	"github.com/mltheuser/ai-router/router/chat"
)

// fallbackMaxTokens is the max_tokens sent when the request sets none and the
// model's own maximum is unknown. 4096 is within every Claude model's limit.
const fallbackMaxTokens = 4096

// --- Anthropic wire types (request) ---

type anthropicChatRequest struct {
	Model        string                 `json:"model"`
	MaxTokens    int                    `json:"max_tokens"`
	System       string                 `json:"system,omitempty"`
	Messages     []anthropicMessage     `json:"messages"`
	Tools        []anthropicTool        `json:"tools,omitempty"`
	Thinking     *anthropicThinking     `json:"thinking,omitempty"`
	OutputConfig *anthropicOutputConfig `json:"output_config,omitempty"`
	CacheControl *anthropicCacheControl `json:"cache_control,omitempty"`
}

type anthropicCacheControl struct {
	Type string `json:"type"`
}

// anthropicMessage carries an ordered list of content blocks. Content is a
// heterogeneous union (text, image, tool_use, tool_result), so the blocks are
// typed structs assembled into an []interface{}.
type anthropicMessage struct {
	Role    string        `json:"role"`
	Content []interface{} `json:"content"`
}

type anthropicTextBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicImageBlock struct {
	Type   string               `json:"type"`
	Source anthropicImageSource `json:"source"`
}

type anthropicImageSource struct {
	Type      string `json:"type"`
	MediaType string `json:"media_type"`
	Data      string `json:"data"`
}

type anthropicToolUseBlock struct {
	Type  string                 `json:"type"`
	ID    string                 `json:"id"`
	Name  string                 `json:"name"`
	Input map[string]interface{} `json:"input"`
}

type anthropicToolResultBlock struct {
	Type      string        `json:"type"`
	ToolUseID string        `json:"tool_use_id"`
	Content   []interface{} `json:"content,omitempty"`
}

type anthropicTool struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	InputSchema map[string]interface{} `json:"input_schema"`
}

type anthropicThinking struct {
	Type string `json:"type"`
	// Display defaults to "omitted", which returns thinking blocks with empty
	// text; "summarized" returns the reasoning trace.
	Display string `json:"display,omitempty"`
}

type anthropicOutputConfig struct {
	Effort string           `json:"effort,omitempty"`
	Format *anthropicFormat `json:"format,omitempty"`
}

type anthropicFormat struct {
	Type   string                 `json:"type"`
	Schema map[string]interface{} `json:"schema,omitempty"`
}

// --- Anthropic wire types (response) ---

type anthropicChatResponse struct {
	ID         string                   `json:"id"`
	Type       string                   `json:"type"`
	Role       string                   `json:"role"`
	Model      string                   `json:"model"`
	StopReason string                   `json:"stop_reason"`
	Content    []anthropicResponseBlock `json:"content"`
	Usage      anthropicUsage           `json:"usage"`
}

// anthropicResponseBlock flattens every block variant the response may contain.
// Input is left raw so tool arguments are decoded once, into a map.
type anthropicResponseBlock struct {
	Type     string          `json:"type"`
	Text     string          `json:"text"`
	Thinking string          `json:"thinking"`
	ID       string          `json:"id"`
	Name     string          `json:"name"`
	Input    json.RawMessage `json:"input"`
}

type anthropicUsage struct {
	InputTokens              int `json:"input_tokens"`
	OutputTokens             int `json:"output_tokens"`
	CacheCreationInputTokens int `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int `json:"cache_read_input_tokens"`
	OutputTokensDetails      struct {
		ThinkingTokens int `json:"thinking_tokens"`
	} `json:"output_tokens_details"`
}

// --- Chat implementation ---

func (p *Provider) Chat(ctx context.Context, req *chat.Request) (*chat.Response, error) {
	// Anthropic requires max_tokens. An explicit value caps the response; when
	// omitted, default to the model's own maximum output so nothing is capped
	// below the model's ceiling.
	var maxTokens int
	if req.MaxTokens != nil {
		maxTokens = *req.MaxTokens
	} else {
		maxTokens = p.modelMaxTokens(ctx, req.Model)
	}

	aReq := toAnthropicRequest(req, maxTokens)

	var aResp anthropicChatResponse
	if err := p.client.Post(ctx, "/messages", aReq, &aResp); err != nil {
		return nil, err
	}

	return mapAnthropicResponse(&aResp), nil
}

// --- Request translation ---

func toAnthropicRequest(req *chat.Request, maxTokens int) *anthropicChatRequest {
	aReq := &anthropicChatRequest{
		Model:     req.Model,
		MaxTokens: maxTokens,
		// Always request automatic prompt caching: it is transparent to
		// clients. Anthropic places the cache breakpoint itself and skips
		// prompts below the model's minimum cacheable length.
		CacheControl: &anthropicCacheControl{Type: "ephemeral"},
	}

	// Anthropic carries the system prompt out of band, not as a message. Pull
	// every system turn into the top-level field and route the rest through
	// appendMessage as user/assistant turns.
	var system []string
	for _, m := range req.Messages {
		if m.Role == chat.RoleSystem {
			if text := chat.TextFromContent(m.Content); text != "" {
				system = append(system, text)
			}
			continue
		}
		appendMessage(aReq, m)
	}
	if len(system) > 0 {
		aReq.System = strings.Join(system, "\n\n")
	}

	// Tools: Anthropic names the schema field input_schema and requires it to be
	// present. Fall back to an empty object schema for argument-less tools.
	for _, t := range req.Tools {
		schema := t.Parameters
		if schema == nil {
			schema = map[string]interface{}{"type": "object"}
		}
		aReq.Tools = append(aReq.Tools, anthropicTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: schema,
		})
	}

	// Reasoning maps to adaptive thinking plus an effort level. "none" omits
	// thinking entirely: some models reject an explicit "disabled".
	if req.ReasoningEffort != nil && *req.ReasoningEffort != chat.ReasoningEffortNone {
		aReq.Thinking = &anthropicThinking{Type: "adaptive", Display: "summarized"}
		aReq.outputConfig().Effort = string(*req.ReasoningEffort)
	}

	// Structured output: Anthropic constrains the response shape via
	// output_config.format rather than a top-level response_format field.
	if rf := req.ResponseFormat; rf != nil && rf.Type == chat.ResponseFormatJSONSchema && rf.JSONSchema != nil {
		aReq.outputConfig().Format = &anthropicFormat{
			Type:   string(chat.ResponseFormatJSONSchema),
			Schema: rf.JSONSchema.Schema,
		}
	}

	// Temperature, top_p and the frequency/presence penalties are never
	// forwarded: some Claude models reject temperature/top_p with HTTP 400, and
	// none supports penalties.

	return aReq
}

// outputConfig lazily initializes the shared output_config block so effort and
// format can both populate it without clobbering each other.
func (r *anthropicChatRequest) outputConfig() *anthropicOutputConfig {
	if r.OutputConfig == nil {
		r.OutputConfig = &anthropicOutputConfig{}
	}
	return r.OutputConfig
}

// appendMessage converts a single shared message to Anthropic content blocks and
// appends them as a turn, coalescing into the previous turn when the role
// matches. Coalescing is required for parallel tool calls: they come back as
// several consecutive tool-result messages that Anthropic expects grouped into
// one user turn.
func appendMessage(aReq *anthropicChatRequest, m chat.Message) {
	var role string
	var blocks []interface{}

	switch m.Role {
	case chat.RoleTool:
		// A tool result is a user turn carrying a tool_result block keyed by the
		// originating tool_use id. Its content is a block list so image parts
		// travel inside the tool_result.
		role = "user"
		blocks = append(blocks, anthropicToolResultBlock{
			Type:      "tool_result",
			ToolUseID: m.ToolCallID,
			Content:   contentBlocks(m.Content),
		})
	case chat.RoleAssistant:
		role = "assistant"
		if text := chat.TextFromContent(m.Content); text != "" {
			blocks = append(blocks, anthropicTextBlock{Type: "text", Text: text})
		}
		for _, tc := range m.ToolCalls {
			input := tc.Function.Arguments
			if input == nil {
				input = map[string]interface{}{}
			}
			blocks = append(blocks, anthropicToolUseBlock{
				Type:  "tool_use",
				ID:    tc.ID,
				Name:  tc.Function.Name,
				Input: input,
			})
		}
	default: // user, plus any unknown role treated as user input
		role = "user"
		blocks = contentBlocks(m.Content)
	}

	if len(blocks) == 0 {
		return
	}

	if n := len(aReq.Messages); n > 0 && aReq.Messages[n-1].Role == role {
		aReq.Messages[n-1].Content = append(aReq.Messages[n-1].Content, blocks...)
		return
	}
	aReq.Messages = append(aReq.Messages, anthropicMessage{Role: role, Content: blocks})
}

// contentBlocks converts shared multimodal content parts to Anthropic blocks.
func contentBlocks(parts []chat.ContentPart) []interface{} {
	blocks := make([]interface{}, 0, len(parts))
	for _, p := range parts {
		switch p.Type {
		case chat.ContentPartText:
			if p.Text == "" {
				continue
			}
			blocks = append(blocks, anthropicTextBlock{Type: "text", Text: p.Text})
		case chat.ContentPartImage:
			blocks = append(blocks, anthropicImageBlock{
				Type: "image",
				Source: anthropicImageSource{
					Type:      "base64",
					MediaType: p.MimeType,
					Data:      p.Base64Data,
				},
			})
		}
	}
	return blocks
}

// --- Response translation ---

func mapAnthropicResponse(aResp *anthropicChatResponse) *chat.Response {
	// input_tokens excludes cached tokens, both reads and writes.
	promptTokens := aResp.Usage.InputTokens + aResp.Usage.CacheReadInputTokens + aResp.Usage.CacheCreationInputTokens
	resp := chat.Response{
		Model: aResp.Model,
		Usage: chat.Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: aResp.Usage.OutputTokens,
			TotalTokens:      promptTokens + aResp.Usage.OutputTokens,
			ReasoningTokens:  aResp.Usage.OutputTokensDetails.ThinkingTokens,
			CacheReadTokens:  aResp.Usage.CacheReadInputTokens,
		},
		Message: chat.Message{Role: chat.RoleAssistant},
	}

	var text, reasoning string
	var toolCalls []chat.ToolCall
	for _, b := range aResp.Content {
		switch b.Type {
		case "text":
			text += b.Text
		case "thinking":
			reasoning += b.Thinking
		case "tool_use":
			toolCalls = append(toolCalls, chat.ToolCall{
				ID: b.ID,
				Function: chat.ToolCallFunction{
					Name:      b.Name,
					Arguments: parseToolInput(b.Input),
				},
			})
		}
	}

	resp.Message.Content = chat.TextContent(text)
	resp.Message.ReasoningContent = reasoning
	resp.Message.ToolCalls = toolCalls

	resp.FinishReason = mapStopReason(aResp.StopReason)

	return &resp
}

// parseToolInput decodes a tool_use input object into a map.
func parseToolInput(raw json.RawMessage) map[string]interface{} {
	args := map[string]interface{}{}
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &args)
	}
	return args
}

// mapStopReason maps Anthropic stop reasons to the shared FinishReason.
func mapStopReason(reason string) chat.FinishReason {
	switch reason {
	case "end_turn", "stop_sequence", "":
		return chat.FinishReasonStop
	case "tool_use":
		return chat.FinishReasonToolCalls
	case "max_tokens", "model_context_window_exceeded":
		return chat.FinishReasonLength
	case "refusal":
		return chat.FinishReasonContentFilter
	default:
		return chat.FinishReason(reason)
	}
}
