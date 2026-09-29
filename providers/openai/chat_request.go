package openai

import (
	"encoding/json"

	"github.com/mltheuser/ai-router/router/chat"
)

// --- Responses wire types (request) ---

type responsesRequest struct {
	Model string `json:"model"`
	// Store is always false: the full history is replayed on every request,
	// so nothing needs to be persisted server-side.
	Store           bool                `json:"store"`
	Input           []interface{}       `json:"input"`
	Tools           []responsesTool     `json:"tools,omitempty"`
	Reasoning       *responsesReasoning `json:"reasoning,omitempty"`
	Text            *responsesText      `json:"text,omitempty"`
	MaxOutputTokens *int                `json:"max_output_tokens,omitempty"`
}

// responsesMessage is a message input item. Input content parts are
// input_text/input_image; assistant history uses output_text.
type responsesMessage struct {
	Type    string                 `json:"type"`
	Role    string                 `json:"role"`
	Content []responsesContentPart `json:"content"`
}

// responsesContentPart is one content part. Exactly one of Text or ImageURL
// is set, matching Type.
type responsesContentPart struct {
	Type     string `json:"type"`
	Text     string `json:"text,omitempty"`
	ImageURL string `json:"image_url,omitempty"`
}

// responsesFunctionCall is an assistant tool call replayed as an input item.
// Arguments are serialized as a JSON string.
type responsesFunctionCall struct {
	Type      string `json:"type"`
	CallID    string `json:"call_id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// responsesFunctionCallOutput carries a tool result, matched to its call by
// CallID. Output is a list of content parts so image parts reach the model.
type responsesFunctionCallOutput struct {
	Type   string                 `json:"type"`
	CallID string                 `json:"call_id"`
	Output []responsesContentPart `json:"output"`
}

type responsesTool struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Parameters  map[string]interface{} `json:"parameters,omitempty"`
	// Strict is always sent as false: when omitted, OpenAI upgrades any
	// compatible schema to strict mode, which makes every parameter required.
	Strict bool `json:"strict"`
}

type responsesReasoning struct {
	Effort  string `json:"effort,omitempty"`
	Summary string `json:"summary"`
}

type responsesText struct {
	Format responsesFormat `json:"format"`
}

type responsesFormat struct {
	Type        string                 `json:"type"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	Schema      map[string]interface{} `json:"schema,omitempty"`
	Strict      bool                   `json:"strict"`
}

// --- Request translation ---

func toResponsesRequest(req *chat.Request) *responsesRequest {
	oReq := &responsesRequest{
		Model:           req.Model,
		Store:           false,
		MaxOutputTokens: req.MaxTokens,
		// Always ask for a concise reasoning summary: it is the only visible
		// reasoning text the API returns ("auto"/"detailed" often come back
		// empty), and it is accepted even when reasoning is off.
		Reasoning: &responsesReasoning{Summary: "concise"},
	}

	if req.ReasoningEffort != nil {
		oReq.Reasoning.Effort = string(*req.ReasoningEffort)
	}

	for _, m := range req.Messages {
		oReq.Input = append(oReq.Input, inputItems(m)...)
	}

	for _, t := range req.Tools {
		oReq.Tools = append(oReq.Tools, responsesTool{
			Type:        "function",
			Name:        t.Name,
			Description: t.Description,
			Parameters:  t.Parameters,
		})
	}

	if rf := req.ResponseFormat; rf != nil && rf.Type == chat.ResponseFormatJSONSchema && rf.JSONSchema != nil {
		oReq.Text = &responsesText{Format: responsesFormat{
			Type:        string(chat.ResponseFormatJSONSchema),
			Name:        rf.JSONSchema.Name,
			Description: rf.JSONSchema.Description,
			Schema:      rf.JSONSchema.Schema,
			Strict:      true,
		}}
	}

	// Temperature, top_p and the frequency/presence penalties are never
	// forwarded: reasoning models reject them with HTTP 400 whenever reasoning
	// is enabled, which is the default.

	return oReq
}

// inputItems converts one shared message to Responses input items.
func inputItems(m chat.Message) []interface{} {
	switch m.Role {
	case chat.RoleSystem:
		return []interface{}{newMessage("developer", inputContent(m.Content))}
	case chat.RoleTool:
		return []interface{}{responsesFunctionCallOutput{
			Type:   "function_call_output",
			CallID: m.ToolCallID,
			Output: inputContent(m.Content),
		}}
	case chat.RoleAssistant:
		// Assistant text replays as a message with output_text; each tool call
		// becomes its own function_call item after it. ReasoningContent is not
		// replayed: the API ignores reasoning it did not issue itself.
		var items []interface{}
		if text := chat.TextFromContent(m.Content); text != "" {
			items = append(items, newMessage("assistant", []responsesContentPart{{Type: "output_text", Text: text}}))
		}
		for _, tc := range m.ToolCalls {
			args := tc.Function.Arguments
			if args == nil {
				args = map[string]interface{}{}
			}
			argsJSON, _ := json.Marshal(args)
			items = append(items, responsesFunctionCall{
				Type:      "function_call",
				CallID:    tc.ID,
				Name:      tc.Function.Name,
				Arguments: string(argsJSON),
			})
		}
		return items
	default: // user, plus any unknown role treated as user input
		return []interface{}{newMessage("user", inputContent(m.Content))}
	}
}

// newMessage builds a message input item with the given role and content.
func newMessage(role string, content []responsesContentPart) responsesMessage {
	return responsesMessage{Type: "message", Role: role, Content: content}
}

// inputContent converts shared multimodal content parts to Responses input
// parts. Images are sent inline as base64 data URLs.
func inputContent(parts []chat.ContentPart) []responsesContentPart {
	result := make([]responsesContentPart, 0, len(parts))
	for _, p := range parts {
		switch p.Type {
		case chat.ContentPartText:
			result = append(result, responsesContentPart{Type: "input_text", Text: p.Text})
		case chat.ContentPartImage:
			result = append(result, responsesContentPart{
				Type:     "input_image",
				ImageURL: "data:" + p.MimeType + ";base64," + p.Base64Data,
			})
		}
	}
	return result
}
