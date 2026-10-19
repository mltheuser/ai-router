package chat

import (
	"encoding/json"
	"strings"
	"testing"
)

// A message without text must encode its content as [], not null: clients
// declare content non-null, and a text part without text is rejected by some
// upstreams when the message is sent back.
func TestTextContentOfEmptyTextEncodesAsEmptyList(t *testing.T) {
	data, err := json.Marshal(Message{Role: RoleAssistant, Content: TextContent("")})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), `"content":[]`) {
		t.Errorf("got %s; want content encoded as []", data)
	}
}
