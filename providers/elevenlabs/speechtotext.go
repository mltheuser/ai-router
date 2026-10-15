package elevenlabs

import (
	"context"

	"github.com/mltheuser/ai-router/httpx"
	"github.com/mltheuser/ai-router/router"
	"github.com/mltheuser/ai-router/router/speechtotext"
)

// speechToTextModels are the speech-to-text models served, by ID, with their
// pay-as-you-go price per hour of audio (https://elevenlabs.io/pricing/api).
// GET /v1/models lists no speech-to-text model, so they are maintained here.
var speechToTextModels = map[string]float64{
	"scribe_v2": 0.22,
}

// --- Speech-to-text wire types ---

type transcriptResponse struct {
	Text string `json:"text"`
}

func (p *Provider) ListSpeechToTextModels(_ context.Context) ([]speechtotext.Model, error) {
	models := make([]speechtotext.Model, 0, len(speechToTextModels))
	for id, cost := range speechToTextModels {
		models = append(models, speechtotext.Model{
			ModelRef:         router.NewModelRef(p, id),
			CostPerAudioHour: &cost,
		})
	}
	return models, nil
}

// SpeechToText transcribes the audio as plain text: audio events such as
// laughter stay out of the transcript, and no timestamps are requested, as
// only the text is returned.
func (p *Provider) SpeechToText(ctx context.Context, model string, wav []byte) (string, error) {
	var form httpx.Form
	form.Field("model_id", model)
	form.File("file", "audio.wav", "audio/wav", wav)
	form.Field("tag_audio_events", "false")
	form.Field("timestamps_granularity", "none")

	var resp transcriptResponse
	if err := p.client.PostForm(ctx, "/v1/speech-to-text", &form, &resp); err != nil {
		return "", err
	}
	return resp.Text, nil
}
