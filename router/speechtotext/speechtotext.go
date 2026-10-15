// Package speechtotext is the speech-to-text use case: speech in, as a WAV
// file, its transcript out.
package speechtotext

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"net/http"

	"github.com/mltheuser/ai-router/httpx"
	"github.com/mltheuser/ai-router/router"
)

// Provider is implemented by every backend that serves speech to text.
type Provider interface {
	router.Provider

	// ListSpeechToTextModels returns the speech-to-text models available at
	// this provider.
	ListSpeechToTextModels(ctx context.Context) ([]Model, error)

	// SpeechToText returns the transcript of wav, a WAV file of 16-bit mono
	// PCM, made by the model with the bare ID model.
	SpeechToText(ctx context.Context, model string, wav []byte) (string, error)
}

// Model describes a speech-to-text model at one provider.
type Model struct {
	router.ModelRef

	// CostPerAudioHour is the price per hour of input audio. Nil means
	// unknown; zero means free.
	CostPerAudioHour *float64 `json:"cost_per_audio_hour,omitempty"`
}

// UseCase serves speech to text.
type UseCase struct {
	*router.Base[Model, Provider]
}

// New builds the speech-to-text use case over the providers that implement
// Provider.
func New(ctx context.Context, providers []router.Provider) *UseCase {
	return &UseCase{router.NewBase(ctx, router.Spec[Model, Provider]{
		Name:      "speech-to-text",
		List:      Provider.ListSpeechToTextModels,
		Prefer:    prefer,
		Scenarios: scenarios,
	}, providers)}
}

// prefer picks the cheaper model.
func prefer(a, b Model) bool {
	return router.LessKnown(a.CostPerAudioHour, b.CostPerAudioHour)
}

// Handle serves POST /v1/speech-to-text.
func (u *UseCase) Handle(w http.ResponseWriter, r *http.Request) error {
	var req Request
	if err := httpx.DecodeJSON(r, &req); err != nil {
		return err
	}
	if req.Audio == "" {
		return httpx.NewError(http.StatusBadRequest, "audio is required")
	}
	wav, err := base64.StdEncoding.DecodeString(req.Audio)
	if err != nil {
		return httpx.NewError(http.StatusBadRequest, "audio is not valid base64: "+err.Error())
	}
	if err := checkWAV(wav); err != nil {
		return httpx.NewError(http.StatusBadRequest, "audio is not a WAV file of 16-bit mono PCM: "+err.Error())
	}

	model, p, err := u.Resolve(req.Model)
	if err != nil {
		return err
	}

	text, err := p.SpeechToText(r.Context(), model.ID, wav)
	if err != nil {
		return fmt.Errorf("%s: %w", p.Name(), err)
	}
	httpx.WriteJSON(w, Response{Model: model.ID, Text: text})
	return nil
}

// checkWAV checks that data is a WAV file of 16-bit mono PCM: a fmt chunk
// declaring that format, followed by a data chunk. A data chunk whose declared
// size overruns the file, as streamed WAVs declare, is taken to run to the end
// of the file.
func checkWAV(data []byte) error {
	if len(data) < 12 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return errors.New("no RIFF/WAVE header")
	}

	hasFmt := false
	for rest := data[12:]; len(rest) >= 8; {
		id, size := string(rest[0:4]), int(binary.LittleEndian.Uint32(rest[4:8]))
		body := rest[8:]
		if size > len(body) {
			if id != "data" {
				return fmt.Errorf("chunk '%s' is truncated", id)
			}
			size = len(body)
		}
		body = body[:size]

		switch id {
		case "fmt ":
			if size < 16 {
				return fmt.Errorf("fmt chunk is %d bytes; expected at least 16", size)
			}
			format := binary.LittleEndian.Uint16(body[0:2])
			channels := binary.LittleEndian.Uint16(body[2:4])
			sampleRate := binary.LittleEndian.Uint32(body[4:8])
			bits := binary.LittleEndian.Uint16(body[14:16])
			if format != 1 || channels != 1 || bits != 16 || sampleRate == 0 {
				return fmt.Errorf("format %d, %d channels, %d bits, %d Hz; expected PCM (1), 1 channel, 16 bits",
					format, channels, bits, sampleRate)
			}
			hasFmt = true
		case "data":
			if !hasFmt {
				return errors.New("data chunk precedes the fmt chunk")
			}
			return nil
		}

		// Chunks are padded to an even size.
		rest = rest[min(8+size+size%2, len(rest)):]
	}
	return errors.New("no data chunk")
}
