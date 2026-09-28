package openai

import (
	"context"

	"github.com/mltheuser/ai-router/usecase/chat"
)

// The shared API is chat-completions style (a list of role-tagged messages),
// but this provider talks to OpenAI's Responses API: on current models it is
// the only endpoint that combines reasoning with function tools, returns
// reasoning summaries, and delivers images inside tool results to the model.
// Responses takes a flat list of input items instead of messages; tool calls
// and tool results are items of their own rather than message fields.
//
// The translation lives in chat_request.go (shared → Responses) and
// chat_response.go (Responses → shared).

// Chat sends a chat request to the OpenAI Responses API and maps the response
// back to the shared API type.
func (p *Provider) Chat(ctx context.Context, req *chat.Request) (*chat.Response, error) {
	oReq := toResponsesRequest(req)

	var oResp responsesResponse
	if err := p.client.post(ctx, "/responses", oReq, &oResp); err != nil {
		return nil, err
	}

	return mapResponsesResponse(&oResp), nil
}
