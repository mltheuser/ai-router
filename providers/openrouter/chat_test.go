package openrouter

import (
	"encoding/json"
	"testing"

	"github.com/mltheuser/ai-router/router/chat"
)

// An assistant reply that only calls tools, sent back as received, must reach
// OpenRouter without a text part: Claude and Gemini reject one without text.
func TestToolOnlyAssistantMessageHasNoTextPart(t *testing.T) {
	reply := chat.Message{
		Role:    chat.RoleAssistant,
		Content: chat.TextContent(""),
		ToolCalls: []chat.ToolCall{{
			ID:       "call_1",
			Function: chat.ToolCallFunction{Name: "get_weather", Arguments: map[string]interface{}{"city": "Berlin"}},
		}},
	}
	orReq := toOpenRouterRequest(&chat.Request{Model: "m", Messages: []chat.Message{reply}})

	msg := orReq.Messages[0]
	if len(msg.Content) != 0 {
		data, _ := json.Marshal(msg.Content)
		t.Errorf("content is %s; want no parts", data)
	}
	if len(msg.ToolCalls) != 1 || msg.ToolCalls[0].ID != "call_1" {
		t.Errorf("tool calls are %+v; want the one call", msg.ToolCalls)
	}
}
