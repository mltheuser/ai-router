package speechtotext

import (
	"context"
	_ "embed"
	"encoding/base64"
	"fmt"
	"slices"
	"strings"
	"unicode"

	"github.com/mltheuser/ai-router/router"
)

// sentenceWAV is sentence spoken, as text-to-speech returns it: a WAV file of
// 16-bit mono PCM at 24 kHz. It was made by POST /v1/text-to-speech with
// ElevenLabs' eleven_multilingual_v2 in the voice "Mark - Natural
// Conversations" (UgBBYS2sOqTuMpoF3BR0).
//
//go:embed resources/sentence.wav
var sentenceWAV []byte

// sentence is what sentenceWAV says. A pangram that every model transcribes
// verbatim leaves no room for a transcript to be both right and different.
const sentence = "The quick brown fox jumps over the lazy dog."

var scenarios = []router.Scenario[Model]{
	{Name: "transcription", Run: runTranscription},
}

// runTranscription verifies that the model transcribes a spoken sentence
// word for word. Case and punctuation are up to the model.
func runTranscription(ctx context.Context, url string, m Model, res *router.Result) {
	resp, err := router.PostJSON[Response](ctx, url, Request{
		Model: m.Model,
		Audio: base64.StdEncoding.EncodeToString(sentenceWAV),
	})
	if err != nil {
		res.Fail("transcription", err.Error())
		return
	}
	if resp.Model != m.ID {
		res.Fail("transcription", fmt.Sprintf("response names model '%s'; expected '%s'", resp.Model, m.ID))
		return
	}
	res.Pass("transcription")

	if got, want := words(resp.Text), words(sentence); !slices.Equal(got, want) {
		res.Fail("transcript", fmt.Sprintf("transcript %q does not say %q", resp.Text, sentence))
	} else {
		res.Pass("transcript")
	}
}

// words splits text into its lowercase words, dropping punctuation.
func words(text string) []string {
	return strings.FieldsFunc(strings.ToLower(text), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}
