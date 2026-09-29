package openai

import (
	"context"

	"github.com/mltheuser/ai-router/router/chat"
)

// The shared API is chat-completions style (a list of role-tagged messages),
// but this provider talks to OpenAI's Responses API: it is the only endpoint
// that combines reasoning with function tools, returns
// reasoning summaries, and delivers images inside tool results to the model.
// Responses takes a flat list of input items instead of messages; tool calls
// and tool results are items of their own rather than message fields.

func (p *Provider) Chat(ctx context.Context, req *chat.Request) (*chat.Response, error) {
	oReq := toResponsesRequest(req)

	var oResp responsesResponse
	if err := p.client.Post(ctx, "/responses", oReq, &oResp); err != nil {
		return nil, err
	}

	return mapResponsesResponse(&oResp), nil
}
