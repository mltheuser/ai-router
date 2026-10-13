package elevenlabs

import (
	"context"
	"fmt"
	"net/url"

	"github.com/mltheuser/ai-router/router/texttospeech"
)

// voicesPageSize is the largest page GET /v2/voices serves.
const voicesPageSize = 100

// --- Voice wire types ---

type voicesResponse struct {
	Voices        []voice `json:"voices"`
	HasMore       bool    `json:"has_more"`
	NextPageToken string  `json:"next_page_token"`
}

type voice struct {
	VoiceID       string        `json:"voice_id"`
	Name          string        `json:"name"`
	SafetyControl string        `json:"safety_control"`
	Sharing       *voiceSharing `json:"sharing"`
}

type voiceSharing struct {
	Status string `json:"status"`
}

// usable reports whether ElevenLabs serves speech in the voice at all: it
// refuses voices it has banned, and library voices whose sharing it disabled.
func (v voice) usable() bool {
	switch v.SafetyControl {
	case "BAN", "ENTERPRISE_BAN":
		return false
	}
	return v.Sharing == nil || v.Sharing.Status != "copied_disabled"
}

// listVoices returns the usable voices of the account's My Voices. Every one
// works with every text-to-speech model: ElevenLabs speaks even a
// professional clone with a model it was not fine-tuned on.
func (p *Provider) listVoices(ctx context.Context) ([]texttospeech.Voice, error) {
	var voices []texttospeech.Voice
	token := ""
	for {
		query := url.Values{}
		query.Set("page_size", fmt.Sprint(voicesPageSize))
		if token != "" {
			query.Set("next_page_token", token)
		}

		var resp voicesResponse
		if err := p.client.Get(ctx, "/v2/voices?"+query.Encode(), &resp); err != nil {
			return nil, fmt.Errorf("listing elevenlabs voices: %w", err)
		}
		for _, v := range resp.Voices {
			if v.usable() {
				voices = append(voices, texttospeech.Voice{ID: v.VoiceID, Name: v.Name})
			}
		}

		// A missing or repeated token ends the listing, so it always ends.
		if !resp.HasMore || resp.NextPageToken == "" || resp.NextPageToken == token {
			return voices, nil
		}
		token = resp.NextPageToken
	}
}
