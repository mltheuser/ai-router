package openai

import (
	"encoding/json"
	"strings"

	"github.com/mltheuser/ai-router/usecase/chat"
)

// --- Responses wire types (response) ---

type responsesResponse struct {
	Model             string                `json:"model"`
	Status            string                `json:"status"`
	IncompleteDetails *responsesIncomplete  `json:"incomplete_details"`
	Output            []responsesOutputItem `json:"output"`
	Usage             responsesUsage        `json:"usage"`
}

type responsesIncomplete struct {
	Reason string `json:"reason"`
}

// responsesOutputItem flattens the output item variants we consume:
// message (Content), reasoning (Summary) and function_call (CallID, Name,
// Arguments).
type responsesOutputItem struct {
	Type      string                      `json:"type"`
	Content   []responsesOutputContent    `json:"content"`
	Summary   []responsesReasoningSummary `json:"summary"`
	CallID    string                      `json:"call_id"`
	Name      string                      `json:"name"`
	Arguments string                      `json:"arguments"`
}

type responsesOutputContent struct {
	Type    string `json:"type"`
	Text    string `json:"text"`
	Refusal string `json:"refusal"`
}

type responsesReasoningSummary struct {
	Text string `json:"text"`
}

// responsesUsage reports token counts. input_tokens already includes cached
// tokens, and output_tokens already includes reasoning tokens.
type responsesUsage struct {
	InputTokens        int `json:"input_tokens"`
	OutputTokens       int `json:"output_tokens"`
	TotalTokens        int `json:"total_tokens"`
	InputTokensDetails struct {
		CachedTokens int `json:"cached_tokens"`
	} `json:"input_tokens_details"`
	OutputTokensDetails struct {
		ReasoningTokens int `json:"reasoning_tokens"`
	} `json:"output_tokens_details"`
}

// --- Response translation ---

func mapResponsesResponse(oResp *responsesResponse) *chat.Response {
	resp := chat.Response{
		Model: oResp.Model,
		Usage: chat.Usage{
			PromptTokens:     oResp.Usage.InputTokens,
			CompletionTokens: oResp.Usage.OutputTokens,
			TotalTokens:      oResp.Usage.TotalTokens,
			ReasoningTokens:  oResp.Usage.OutputTokensDetails.ReasoningTokens,
			CacheReadTokens:  oResp.Usage.InputTokensDetails.CachedTokens,
		},
		Message: chat.Message{Role: chat.RoleAssistant},
	}

	// The output is a list of items: message items carry the answer text (a
	// tool-calling turn may add a short preamble message), reasoning items carry
	// summaries, and each function_call item is one tool call.
	var texts, summaries []string
	refused := false
	for _, item := range oResp.Output {
		switch item.Type {
		case "message":
			text, isRefusal := messageText(item.Content)
			if text != "" {
				texts = append(texts, text)
			}
			refused = refused || isRefusal
		case "reasoning":
			for _, s := range item.Summary {
				if s.Text != "" {
					summaries = append(summaries, s.Text)
				}
			}
		case "function_call":
			resp.Message.ToolCalls = append(resp.Message.ToolCalls, chat.ToolCall{
				ID: item.CallID,
				Function: chat.ToolCallFunction{
					Name:      item.Name,
					Arguments: parseArguments(item.Arguments),
				},
			})
		}
	}

	resp.Message.Content = chat.TextContent(strings.Join(texts, "\n\n"))
	resp.Message.ReasoningContent = strings.Join(summaries, "\n\n")
	resp.FinishReason = mapFinishReason(oResp, len(resp.Message.ToolCalls) > 0, refused)

	return &resp
}

// messageText concatenates the text of a message item's content parts. A
// refusal part carries its explanation as text and is reported separately.
func messageText(content []responsesOutputContent) (text string, refused bool) {
	for _, c := range content {
		switch c.Type {
		case "output_text":
			text += c.Text
		case "refusal":
			text += c.Refusal
			refused = true
		}
	}
	return text, refused
}

// parseArguments converts a JSON-encoded arguments string to a map.
func parseArguments(raw string) map[string]interface{} {
	var args map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &args); err != nil {
		return map[string]interface{}{"raw": raw}
	}
	return args
}

// mapFinishReason derives the shared FinishReason. The Responses API has no
// finish reason: it reports a status (plus a reason when incomplete), and tool
// calls are signalled by function_call items. The type is passthrough-friendly,
// so unrecognized values are forwarded unchanged.
func mapFinishReason(oResp *responsesResponse, hasToolCalls, refused bool) chat.FinishReason {
	var incompleteReason string
	if oResp.IncompleteDetails != nil {
		incompleteReason = oResp.IncompleteDetails.Reason
	}

	switch {
	case oResp.Status == "incomplete" && incompleteReason == "max_output_tokens":
		return chat.FinishReasonLength
	case oResp.Status == "incomplete" && incompleteReason == "content_filter":
		return chat.FinishReasonContentFilter
	case oResp.Status == "incomplete":
		return chat.FinishReason(incompleteReason)
	case oResp.Status != "completed" && oResp.Status != "":
		return chat.FinishReason(oResp.Status)
	case hasToolCalls:
		return chat.FinishReasonToolCalls
	case refused:
		return chat.FinishReasonContentFilter
	default:
		return chat.FinishReasonStop
	}
}
