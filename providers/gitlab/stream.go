package gitlab

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/mltheuser/ai-router/httpx"
)

// Cloudflare, in front of the gateway, gives up on a response that sends
// nothing for about 125 seconds and answers 524, well before the gateway's own
// 240-second limit. A long generation sent as one response is cut at 125
// seconds. Streamed, it sends bytes from the start and runs until the
// gateway's limit. So the provider always streams from the gateway, reads the
// stream to its end and rebuilds the message the copied mapping expects;
// clients still get one response.

// streamingRequest is the request with stream set. Embedding keeps the copied
// request type unchanged: its fields are marshaled into the same object.
type streamingRequest struct {
	*anthropicChatRequest
	Stream bool `json:"stream"`
}

// streamEvent is one server-sent event of a streamed Messages response,
// flattened over every event type.
type streamEvent struct {
	Type         string                  `json:"type"`
	Message      *anthropicChatResponse  `json:"message"`
	Index        int                     `json:"index"`
	ContentBlock *anthropicResponseBlock `json:"content_block"`
	Delta        streamDelta             `json:"delta"`
	Usage        json.RawMessage         `json:"usage"`
	Error        json.RawMessage         `json:"error"`
}

// streamDelta flattens the delta of content_block_delta and message_delta.
type streamDelta struct {
	Type        string `json:"type"`
	Text        string `json:"text"`
	Thinking    string `json:"thinking"`
	PartialJSON string `json:"partial_json"`
	StopReason  string `json:"stop_reason"`
}

// postMessages sends aReq to /messages as a stream and returns the message the
// stream adds up to.
func postMessages(ctx context.Context, client *httpx.Client, aReq *anthropicChatRequest) (*anthropicChatResponse, error) {
	body, err := client.PostBytes(ctx, "/messages", streamingRequest{anthropicChatRequest: aReq, Stream: true})
	if errors.Is(err, io.ErrUnexpectedEOF) {
		return nil, cutOff()
	}
	if err != nil {
		return nil, err
	}
	return assemble(body)
}

// assemble rebuilds the message from the stream's events, in the shape of a
// non-streamed response.
func assemble(body []byte) (*anthropicChatResponse, error) {
	var (
		msg      *anthropicChatResponse
		inputs   = map[int]*strings.Builder{}
		complete bool
	)

	for data := range events(body) {
		var ev streamEvent
		if err := json.Unmarshal(data, &ev); err != nil {
			return nil, fmt.Errorf("parsing stream event: %w", err)
		}

		switch ev.Type {
		case "message_start":
			if ev.Message == nil {
				return nil, fmt.Errorf("parsing stream: message_start without a message")
			}
			msg = ev.Message
			msg.Content = nil
		case "error":
			return nil, httpx.NewError(http.StatusBadGateway, fmt.Sprintf("API error (in stream): %s", ev.Error))
		case "message_stop":
			complete = true
		}
		if msg == nil {
			continue
		}

		switch ev.Type {
		case "content_block_start":
			if ev.ContentBlock == nil || ev.Index != len(msg.Content) {
				return nil, fmt.Errorf("parsing stream: content block %d starts out of order", ev.Index)
			}
			msg.Content = append(msg.Content, *ev.ContentBlock)
		case "content_block_delta":
			if ev.Index < 0 || ev.Index >= len(msg.Content) {
				return nil, fmt.Errorf("parsing stream: delta for unknown content block %d", ev.Index)
			}
			block := &msg.Content[ev.Index]
			switch ev.Delta.Type {
			case "text_delta":
				block.Text += ev.Delta.Text
			case "thinking_delta":
				block.Thinking += ev.Delta.Thinking
			case "input_json_delta":
				if inputs[ev.Index] == nil {
					inputs[ev.Index] = &strings.Builder{}
				}
				inputs[ev.Index].WriteString(ev.Delta.PartialJSON)
			}
		case "message_delta":
			msg.StopReason = ev.Delta.StopReason
			// The usage here is cumulative. Unmarshaling onto the usage from
			// message_start overwrites only the fields the event carries.
			if len(ev.Usage) > 0 {
				if err := json.Unmarshal(ev.Usage, &msg.Usage); err != nil {
					return nil, fmt.Errorf("parsing stream usage: %w", err)
				}
			}
		}
	}

	if msg == nil || !complete {
		return nil, cutOff()
	}
	// A tool call's input arrives as JSON fragments; the start event carries
	// only an empty object.
	for i, input := range inputs {
		msg.Content[i].Input = json.RawMessage(input.String())
	}
	return msg, nil
}

// events yields the data of each server-sent event in body.
func events(body []byte) func(yield func([]byte) bool) {
	return func(yield func([]byte) bool) {
		scanner := bufio.NewScanner(bytes.NewReader(body))
		scanner.Buffer(nil, len(body)+1)
		var data []byte
		for scanner.Scan() {
			line := scanner.Bytes()
			switch {
			case len(line) == 0:
				if len(data) > 0 && !yield(data) {
					return
				}
				data = nil
			case bytes.HasPrefix(line, []byte("data:")):
				if len(data) > 0 {
					data = append(data, '\n')
				}
				data = append(data, bytes.TrimPrefix(bytes.TrimPrefix(line, []byte("data:")), []byte(" "))...)
			}
		}
		if len(data) > 0 {
			yield(data)
		}
	}
}

// cutOff is the error for a stream that ended before the message did: the
// gateway closes a request at its 240-second limit without an error event.
func cutOff() error {
	return httpx.NewError(http.StatusGatewayTimeout,
		"the GitLab gateway ended the response before it was complete; it allows a request at most 240 seconds")
}
