package api

// Message roles.
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

// Reasoning effort levels.
const (
	ReasoningEffortNone   = "none"
	ReasoningEffortLow    = "low"
	ReasoningEffortMedium = "medium"
	ReasoningEffortHigh   = "high"
)

// ChatRequest represents a chat completion request.
type ChatRequest struct {
	Model            string           `json:"model"`
	Messages         []ChatMessage    `json:"messages"`
	// Generation parameters
	FrequencyPenalty *float64         `json:"frequency_penalty,omitempty"`
	MaxTokens        *int             `json:"max_tokens,omitempty"`
	PresencePenalty  *float64         `json:"presence_penalty,omitempty"`
	Temperature      *float64         `json:"temperature,omitempty"`
	TopP             *float64         `json:"top_p,omitempty"`
	// Structured Output
	ResponseFormat   *ResponseFormat  `json:"response_format,omitempty"`
	// Reasoning
	ReasoningEffort  *string          `json:"reasoning_effort,omitempty"`
	// Tool Calling
	Tools            []ToolDefinition `json:"tools,omitempty"`
}

// ResponseFormatType represents the type of response format.
type ResponseFormatType string

const (
	ResponseFormatJSONSchema ResponseFormatType = "json_schema"
)

// ResponseFormat specifies the format of the model's response.
type ResponseFormat struct {
	Type       ResponseFormatType `json:"type"`
	JSONSchema *JSONSchema        `json:"json_schema"`
}

// JSONSchema defines the schema for structured JSON output.
type JSONSchema struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
}

// ToolDefinition describes a tool the model may call.
// Providers wrap this in their own wire format (e.g. {"type":"function","function":{...}}).
type ToolDefinition struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
}

// ToolCall represents a tool invocation requested by the assistant.
// ID is used for matching tool results to calls in parallel tool calling.
// Also used on tool-result messages (role=tool) to identify which call the result answers.
type ToolCall struct {
	ID       string           `json:"id"`
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction is the function invocation inside a ToolCall.
type ToolCallFunction struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ContentPartType discriminates content parts in a multimodal message.
type ContentPartType string

const (
	ContentPartText  ContentPartType = "text"
	ContentPartImage ContentPartType = "image"
)

// ContentPart is one piece of a multimodal message.
type ContentPart struct {
	Type       ContentPartType `json:"type"`
	Text       string          `json:"text,omitempty"`        // for type=text
	MimeType   string          `json:"mime_type,omitempty"`   // for type=image (e.g. "image/png", "image/jpeg", "image/webp", "image/gif")
	Base64Data string          `json:"base64_data,omitempty"` // for type=image (raw base64, no data-URI prefix)
}

// TextContent creates a text-only content slice (convenience for the common case).
func TextContent(text string) []ContentPart {
	return []ContentPart{{Type: ContentPartText, Text: text}}
}

// TextFromContent extracts and concatenates all text parts from content.
func TextFromContent(parts []ContentPart) string {
	var text string
	for _, p := range parts {
		if p.Type == ContentPartText {
			text += p.Text
		}
	}
	return text
}

// ImagesFromContent extracts all base64 image data strings from content.
func ImagesFromContent(parts []ContentPart) []string {
	var images []string
	for _, p := range parts {
		if p.Type == ContentPartImage {
			images = append(images, p.Base64Data)
		}
	}
	return images
}

// ChatMessage represents a message in a chat conversation.
// Different roles use different subsets of fields:
//   - user/system:  Role + Content (text-only or multimodal)
//   - assistant:    Role + Content (+ ReasoningContent, ToolCalls)
//   - tool:         Role + Content + ToolCallID
type ChatMessage struct {
	Role             string        `json:"role"`
	Content          []ContentPart `json:"content"`
	ReasoningContent string        `json:"reasoning_content,omitempty"`
	ToolCalls        []ToolCall    `json:"tool_calls,omitempty"`   // assistant only
	ToolCallID       string        `json:"tool_call_id,omitempty"` // tool result only
}

// ChatResponse represents a chat completion response.
type ChatResponse struct {
	Model  string     `json:"model"`
	Choice ChatChoice `json:"choices"`
	Usage  ChatUsage  `json:"usage"`
}

// ChatChoice represents a single choice in a chat completion response.
type ChatChoice struct {
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// ChatUsage represents token usage for a chat completion request.
type ChatUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
	ReasoningTokens  int `json:"reasoning_tokens,omitempty"`
}
