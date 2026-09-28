package chat

import (
	"context"
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mltheuser/ai-router/router"
)

// brindlemarkGuide is a large (>4096 token) original document about an
// invented nation. It is sent verbatim on both turns of the multi-turn
// scenario so the repeated prefix can trigger a provider-side prompt cache
// read on turn 2. The invented proper nouns (capital "Velmoria", river
// "Quillsong") cannot be answered from training data, making recall
// meaningful and substring-checkable.
//
//go:embed resources/brindlemark_guide.md
var brindlemarkGuide string

//go:embed resources/apple.png
var appleImage []byte

var appleImageBase64 = base64.StdEncoding.EncodeToString(appleImage)

var scenarios = []router.Scenario[Model]{
	{Name: "multi_turn", Run: runMultiTurn},
	{Name: "vision", Applies: has(FeatureVision), Run: runVision},
	{Name: "structured_output", Applies: has(FeatureStructuredOutput), Run: runStructuredOutput},
	{
		Name:    "reasoning",
		Applies: has(FeatureReasoning),
		// High-effort traces can stream well past the default budget
		// (observed 90s+ on some models via cloud providers).
		Timeout: 3 * time.Minute,
		Run:     runReasoning,
	},
	{Name: "tool_calling", Applies: has(FeatureTools), Run: runToolCalling},
	{Name: "tool_result_vision", Applies: has(FeatureTools, FeatureVision), Run: runToolResultVision},
}

func has(features ...Feature) func(Model) bool {
	return func(m Model) bool { return m.Has(features...) }
}

func post(ctx context.Context, url string, req Request) (*Response, error) {
	return router.PostJSON[Response](ctx, url, req)
}

// runMultiTurn verifies multi-turn recall over a large document and observes
// prompt-cache reads.
func runMultiTurn(ctx context.Context, url, model string, res *router.Result) {
	temperature := 0.7
	messages := []Message{{Role: RoleUser, Content: TextContent(
		brindlemarkGuide + "\n\nUsing only the travel guide above, what is the capital city of Brindlemark? Answer concisely.")}}

	resp, err := post(ctx, url, Request{Model: model, Temperature: &temperature, Messages: messages})
	if err != nil {
		res.Fail("single turn chat", err.Error())
		return
	}
	if TextFromContent(resp.Message.Content) == "" {
		res.Fail("single turn chat", "response content is empty")
		return
	}
	res.Pass("single turn chat")

	messages = append(messages, resp.Message, Message{
		Role:    RoleUser,
		Content: TextContent("And which river runs through that city? Answer concisely."),
	})
	resp, err = post(ctx, url, Request{Model: model, Temperature: &temperature, Messages: messages})
	if err != nil {
		res.Fail("multi-turn context recall", err.Error())
		return
	}
	content := TextFromContent(resp.Message.Content)
	if !strings.Contains(strings.ToLower(content), "quillsong") {
		res.Fail("multi-turn context recall", fmt.Sprintf("response did not contain 'Quillsong'. Response: %s", content))
		return
	}
	res.Pass("multi-turn context recall")

	// The repeated document prefix across the two turns should produce a cache read.
	if resp.Usage.CacheReadTokens > 0 {
		res.Pass(fmt.Sprintf("prompt cache read observed (%d tokens)", resp.Usage.CacheReadTokens))
	} else {
		res.Fail("prompt cache read", "no cache read observed — verify the provider/model supports prompt caching")
	}
}

// runVision verifies that the model can describe an image.
func runVision(ctx context.Context, url, model string, res *router.Result) {
	resp, err := post(ctx, url, Request{Model: model, Messages: []Message{{
		Role: RoleUser,
		Content: []ContentPart{
			{Type: ContentPartText, Text: "What fruit do you see in the image? Be concise."},
			{Type: ContentPartImage, MimeType: "image/png", Base64Data: appleImageBase64},
		},
	}}})
	if err != nil {
		res.Fail("image description", err.Error())
		return
	}
	content := TextFromContent(resp.Message.Content)
	if !strings.Contains(strings.ToLower(content), "apple") {
		res.Fail("image description", fmt.Sprintf("response does not contain 'apple': %s", content))
		return
	}
	res.Pass("image description")
}

// runStructuredOutput verifies that the response follows a JSON schema.
func runStructuredOutput(ctx context.Context, url, model string, res *router.Result) {
	schema := map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"location":    map[string]interface{}{"type": "string", "description": "the location"},
			"temperature": map[string]interface{}{"type": "number"},
			"unit":        map[string]interface{}{"type": "string", "enum": []string{"celsius", "fahrenheit"}},
		},
		"required":             []string{"location", "temperature", "unit"},
		"additionalProperties": false,
	}

	resp, err := post(ctx, url, Request{
		Model:    model,
		Messages: []Message{{Role: RoleUser, Content: TextContent("It is 25 degrees celsius in Paris.")}},
		ResponseFormat: &ResponseFormat{
			Type:       ResponseFormatJSONSchema,
			JSONSchema: &JSONSchema{Name: "weather_response", Schema: schema},
		},
	})
	if err != nil {
		res.Fail("structured JSON output", err.Error())
		return
	}

	content := TextFromContent(resp.Message.Content)
	var weather struct {
		Location    string  `json:"location"`
		Temperature float64 `json:"temperature"`
		Unit        string  `json:"unit"`
	}
	if err := json.Unmarshal([]byte(content), &weather); err != nil {
		res.Fail("structured JSON output", fmt.Sprintf("response is not the requested JSON: %v. Content: %s", err, content))
		return
	}
	if weather.Location == "" || weather.Unit == "" {
		res.Fail("structured JSON output", fmt.Sprintf("response misses required fields. Content: %s", content))
		return
	}
	res.Pass("structured JSON output")
}

// runReasoning verifies that the model returns a reasoning trace.
func runReasoning(ctx context.Context, url, model string, res *router.Result) {
	// The prompt must be a NOVEL constraint puzzle, not a canonical
	// brain-teaser: models with adaptive thinking (e.g. Anthropic) answer
	// famous problems without emitting a reasoning trace, whereas a puzzle
	// they can't pattern-match to a memorized answer reliably engages thinking.
	effort := ReasoningEffortHigh
	resp, err := post(ctx, url, Request{
		Model: model,
		Messages: []Message{{Role: RoleUser, Content: TextContent(
			"Three friends - Ana, Ben, and Cy - each have a different pet (cat, dog, fish) and live in houses 1, 2, 3. " +
				"Ana is not in house 1. The dog owner is in house 2. Ben owns the fish. Cy is not in house 3. " +
				"Who owns the cat and in which house? Reason step by step.")}},
		ReasoningEffort: &effort,
	})
	if err != nil {
		res.Fail("reasoning trace present", err.Error())
		return
	}
	if resp.Message.ReasoningContent == "" {
		res.Fail("reasoning trace present", fmt.Sprintf("expected reasoning content, got none. Content was: %s", TextFromContent(resp.Message.Content)))
		return
	}
	res.Pass("reasoning trace present")
}

var arithmeticTools = []ToolDefinition{
	{Name: "add", Description: "Add two integers and return the sum", Parameters: twoIntegers},
	{Name: "multiply", Description: "Multiply two integers and return the product", Parameters: twoIntegers},
}

var twoIntegers = map[string]interface{}{
	"type":     "object",
	"required": []string{"a", "b"},
	"properties": map[string]interface{}{
		"a": map[string]interface{}{"type": "integer", "description": "First number"},
		"b": map[string]interface{}{"type": "integer", "description": "Second number"},
	},
}

// runToolCalling verifies single and parallel tool calling: the model invokes
// tools and incorporates their results.
func runToolCalling(ctx context.Context, url, model string, res *router.Result) {
	results := map[string]string{"add": "5", "multiply": "6"}

	// A parallel-capable model calls both tools at once.
	messages := []Message{{Role: RoleUser, Content: TextContent(
		"What is 2 + 3 and 2 * 3? You must use the add tool and the multiply tool to compute this.")}}
	resp, err := post(ctx, url, Request{Model: model, Messages: messages, Tools: arithmeticTools})
	if err != nil {
		res.Fail("tool invocation", err.Error())
		return
	}
	if resp.FinishReason != FinishReasonToolCalls || len(resp.Message.ToolCalls) == 0 {
		res.Fail("tool invocation", fmt.Sprintf("expected tool calls, got finish_reason '%s'", resp.FinishReason))
		return
	}

	called := map[string]bool{}
	for _, tc := range resp.Message.ToolCalls {
		called[tc.Function.Name] = true
	}
	switch {
	case called["add"] && called["multiply"]:
		res.Pass("parallel tool calling")
	case called["add"] || called["multiply"]:
		res.Fail("parallel tool calling", "the model called only one of the two tools at once")
	default:
		var names []string
		for _, tc := range resp.Message.ToolCalls {
			names = append(names, tc.Function.Name)
		}
		res.Fail("tool invocation", fmt.Sprintf("no known tool called; got %v", names))
		return
	}
	res.Pass("tool invocation")

	// Feed the results back. A model that calls the tools one at a time
	// needs a second round.
	for range 2 {
		messages = append(messages, resp.Message)
		for _, tc := range resp.Message.ToolCalls {
			if result, ok := results[tc.Function.Name]; ok {
				messages = append(messages, Message{Role: RoleTool, Content: TextContent(result), ToolCallID: tc.ID})
			}
		}
		resp, err = post(ctx, url, Request{Model: model, Messages: messages, Tools: arithmeticTools})
		if err != nil {
			res.Fail("tool result incorporation", err.Error())
			return
		}
		if resp.FinishReason != FinishReasonToolCalls {
			break
		}
	}

	content := TextFromContent(resp.Message.Content)
	if !strings.Contains(content, "5") || !strings.Contains(content, "6") {
		res.Fail("tool result incorporation", fmt.Sprintf("final answer lacks the results 5 and 6. Response: %s", content))
		return
	}
	res.Pass("tool result incorporation")
}

// runToolResultVision verifies that an image inside a tool result reaches the
// model.
func runToolResultVision(ctx context.Context, url, model string, res *router.Result) {
	tools := []ToolDefinition{{
		Name:        "take_photo",
		Description: "Take a photo with the camera and return it as an image",
		Parameters:  map[string]interface{}{"type": "object", "properties": map[string]interface{}{}},
	}}

	// The question must not hint at the photo's content: asking "what fruit"
	// lets a model that never received the image guess "apple" and pass.
	messages := []Message{{Role: RoleUser, Content: TextContent(
		"Use the take_photo tool, then describe what the photo shows. Be concise.")}}
	resp, err := post(ctx, url, Request{Model: model, Messages: messages, Tools: tools})
	if err != nil {
		res.Fail("tool invocation", err.Error())
		return
	}
	if resp.FinishReason != FinishReasonToolCalls || len(resp.Message.ToolCalls) == 0 {
		res.Fail("tool invocation", fmt.Sprintf("expected a tool call, got finish_reason '%s'", resp.FinishReason))
		return
	}
	res.Pass("tool invocation")

	messages = append(messages, resp.Message)
	for _, tc := range resp.Message.ToolCalls {
		messages = append(messages, Message{
			Role: RoleTool,
			Content: []ContentPart{
				{Type: ContentPartText, Text: "Photo taken:"},
				{Type: ContentPartImage, MimeType: "image/png", Base64Data: appleImageBase64},
			},
			ToolCallID: tc.ID,
		})
	}
	resp, err = post(ctx, url, Request{Model: model, Messages: messages, Tools: tools})
	if err != nil {
		res.Fail("tool result image incorporation", err.Error())
		return
	}
	content := TextFromContent(resp.Message.Content)
	if !strings.Contains(strings.ToLower(content), "apple") {
		res.Fail("tool result image incorporation", fmt.Sprintf("response does not contain 'apple': %s", content))
		return
	}
	res.Pass("tool result image incorporation")
}
