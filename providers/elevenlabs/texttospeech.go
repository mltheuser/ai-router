package elevenlabs

import (
	"cmp"
	"context"
	"fmt"
	"net/url"
	"slices"

	"github.com/mltheuser/ai-router/router"
	"github.com/mltheuser/ai-router/router/texttospeech"
)

// outputFormat is 16-bit mono PCM WAV at 24 kHz, the usual rate for speech.
// Pay-as-you-go keys get no WAV at 44.1 kHz: ElevenLabs reserves it for its
// Pro subscription.
const outputFormat = "wav_24000"

// costPerMCharsAtRate1 is the pay-as-you-go price per million characters of
// a model whose character cost multiplier is 1: $0.08 per 1,000 characters
// (https://elevenlabs.io/pricing/api).
const costPerMCharsAtRate1 = 80.0

// --- Model wire types ---

type model struct {
	ModelID             string     `json:"model_id"`
	CanDoTextToSpeech   bool       `json:"can_do_text_to_speech"`
	RequiresAlphaAccess bool       `json:"requires_alpha_access"`
	MaximumTextLength   int        `json:"maximum_text_length_per_request"`
	ModelRates          modelRates `json:"model_rates"`
}

type modelRates struct {
	CharacterCostMultiplier float64 `json:"character_cost_multiplier"`
	// CostDiscountMultiplier defaults to 1 (no discount) when absent.
	CostDiscountMultiplier *float64 `json:"cost_discount_multiplier"`
}

func (r modelRates) costPerMChars() float64 {
	discount := 1.0
	if r.CostDiscountMultiplier != nil {
		discount = *r.CostDiscountMultiplier
	}
	return costPerMCharsAtRate1 * r.CharacterCostMultiplier * discount
}

// --- Text-to-speech wire types ---

type speechRequest struct {
	Text    string `json:"text"`
	ModelID string `json:"model_id"`
}

// ListTextToSpeechModels returns the text-to-speech models open to every
// account, each with every usable voice of the account's My Voices.
func (p *Provider) ListTextToSpeechModels(ctx context.Context) ([]texttospeech.Model, error) {
	var models []model
	if err := p.client.Get(ctx, "/v1/models", &models); err != nil {
		return nil, fmt.Errorf("listing elevenlabs models: %w", err)
	}
	voices, err := p.listVoices(ctx)
	if err != nil {
		return nil, err
	}
	if len(voices) == 0 {
		return nil, nil
	}
	slices.SortFunc(voices, func(a, b texttospeech.Voice) int { return cmp.Compare(a.Name, b.Name) })

	var result []texttospeech.Model
	for _, m := range models {
		if !m.CanDoTextToSpeech || m.RequiresAlphaAccess {
			continue
		}
		cost := m.ModelRates.costPerMChars()
		result = append(result, texttospeech.Model{
			ModelRef:      router.NewModelRef(p, m.ModelID),
			Voices:        voices,
			MaxInputChars: m.MaximumTextLength,
			CostPerMChars: &cost,
		})
	}
	return result, nil
}

// TextToSpeech speaks the text in the voice and returns ElevenLabs' WAV file
// as is.
func (p *Provider) TextToSpeech(ctx context.Context, req *texttospeech.Request) ([]byte, error) {
	path := "/v1/text-to-speech/" + url.PathEscape(req.Voice) + "?output_format=" + outputFormat
	return p.client.PostBytes(ctx, path, speechRequest{Text: req.Text, ModelID: req.Model})
}
