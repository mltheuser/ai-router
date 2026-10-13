// Package texttospeech is the text-to-speech use case: a text and a voice in,
// the text spoken in that voice out, as a WAV file.
package texttospeech

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/mltheuser/ai-router/httpx"
	"github.com/mltheuser/ai-router/router"
)

// Provider is implemented by every backend that serves text to speech.
type Provider interface {
	router.Provider

	// ListTextToSpeechModels returns the text-to-speech models available at
	// this provider, each with the voices usable with it. A model without a
	// usable voice is left out.
	ListTextToSpeechModels(ctx context.Context) ([]Model, error)

	// TextToSpeech returns req.Text spoken in req.Voice as a WAV file of 16-bit
	// mono PCM. req.Model is the bare model ID, and req.Voice one of the voices
	// ListTextToSpeechModels lists for it.
	TextToSpeech(ctx context.Context, req *Request) ([]byte, error)
}

// Voice is a voice a model speaks in.
type Voice struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// Model describes a text-to-speech model at one provider, with the voices
// requests to it may name.
type Model struct {
	router.ModelRef
	Voices []Voice `json:"voices"`

	// MaxInputChars is the longest text, in characters, the model accepts per
	// request. Zero means unknown.
	MaxInputChars int `json:"max_input_chars,omitempty"`

	// CostPerMChars is the price per million input characters. Nil means
	// unknown; zero means free.
	CostPerMChars *float64 `json:"cost_per_m_chars,omitempty"`
}

// hasVoice reports whether the model lists the voice with the given ID.
func (m Model) hasVoice(id string) bool {
	return slices.ContainsFunc(m.Voices, func(v Voice) bool { return v.ID == id })
}

// UseCase serves text to speech.
type UseCase struct {
	*router.Base[Model, Provider]
}

// New builds the text-to-speech use case over the providers that implement
// Provider.
func New(ctx context.Context, providers []router.Provider) *UseCase {
	return &UseCase{router.NewBase(ctx, router.Spec[Model, Provider]{
		Name:      "text-to-speech",
		List:      Provider.ListTextToSpeechModels,
		Prefer:    prefer,
		Scenarios: scenarios,
	}, providers)}
}

// prefer picks the cheaper model.
func prefer(a, b Model) bool {
	return router.LessKnown(a.CostPerMChars, b.CostPerMChars)
}

// Handle serves POST /v1/text-to-speech.
func (u *UseCase) Handle(w http.ResponseWriter, r *http.Request) error {
	var req Request
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return err
	}
	if strings.TrimSpace(req.Text) == "" {
		return httpx.NewError(http.StatusBadRequest, "text is required")
	}
	if req.Voice == "" {
		return httpx.NewError(http.StatusBadRequest, "voice is required")
	}

	model, p, err := u.Resolve(req.Model)
	if err != nil {
		return err
	}
	if !model.hasVoice(req.Voice) {
		return httpx.NewError(http.StatusBadRequest, fmt.Sprintf(
			"voice '%s' is not available for model '%s'; see GET /v1/text-to-speech/models", req.Voice, req.Model))
	}
	req.Model = model.ID

	wav, err := p.TextToSpeech(r.Context(), &req)
	if err != nil {
		return fmt.Errorf("%s: %w", p.Name(), err)
	}
	httpx.WriteJSON(w, Response{Model: model.ID, Audio: base64.StdEncoding.EncodeToString(wav)})
	return nil
}
