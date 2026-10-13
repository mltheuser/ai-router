package texttospeech

import (
	"context"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/mltheuser/ai-router/router"
)

var scenarios = []router.Scenario[Model]{
	{Name: "speech", Run: runSpeech},
}

// sentence is long enough to take seconds to speak and short enough to cost
// next to nothing.
const sentence = "The quick brown fox jumps over the lazy dog."

const (
	minDuration = 500 * time.Millisecond
	maxDuration = 20 * time.Second
	// minRMS is the loudness below which audio counts as silence, in 16-bit
	// sample units (full scale is 32767). Speech measures in the thousands.
	minRMS = 100
)

// runSpeech verifies that the model speaks a sentence in its first listed
// voice, and that the audio keeps the response's promise: a WAV file of
// 16-bit mono PCM, of plausible length, not silent.
func runSpeech(ctx context.Context, url string, m Model, res *router.Result) {
	if len(m.Voices) == 0 {
		res.Fail("speech", "the model lists no voices")
		return
	}
	voice := m.Voices[0]
	resp, err := router.PostJSON[Response](ctx, url, Request{Model: m.Model, Voice: voice.ID, Text: sentence})
	if err != nil {
		res.Fail("speech", err.Error())
		return
	}
	data, err := base64.StdEncoding.DecodeString(resp.Audio)
	if err != nil {
		res.Fail("speech", "audio is not valid base64: "+err.Error())
		return
	}
	res.Pass(fmt.Sprintf("speech (voice '%s')", voice.Name))

	w, err := parseWAV(data)
	if err != nil {
		res.Fail("wav format", err.Error())
		return
	}
	res.Pass(fmt.Sprintf("wav format (%d Hz)", w.sampleRate))

	if d := w.duration(); d < minDuration || d > maxDuration {
		res.Fail("duration", fmt.Sprintf("%s of audio for a %d-character sentence; expected %s to %s",
			d, len(sentence), minDuration, maxDuration))
	} else {
		res.Pass(fmt.Sprintf("duration (%s)", d.Round(10*time.Millisecond)))
	}

	if rms := w.rms(); rms < minRMS {
		res.Fail("not silent", fmt.Sprintf("RMS loudness %.0f is below %d", rms, minRMS))
	} else {
		res.Pass("not silent")
	}
}

// wav is the audio of a WAV file of 16-bit mono PCM.
type wav struct {
	sampleRate int
	samples    []byte // little-endian 16-bit samples
}

func (w wav) duration() time.Duration {
	return time.Duration(len(w.samples)/2) * time.Second / time.Duration(w.sampleRate)
}

func (w wav) rms() float64 {
	n := len(w.samples) / 2
	if n == 0 {
		return 0
	}
	var sum float64
	for i := range n {
		s := float64(int16(binary.LittleEndian.Uint16(w.samples[2*i:])))
		sum += s * s
	}
	return math.Sqrt(sum / float64(n))
}

// parseWAV reads a WAV file and checks that it holds 16-bit mono PCM. A data
// chunk whose declared size overruns the file, as streamed WAVs declare, is
// taken to run to the end of the file.
func parseWAV(data []byte) (wav, error) {
	if len(data) < 12 || string(data[0:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return wav{}, errors.New("not a WAV file: no RIFF/WAVE header")
	}

	var (
		w      wav
		hasFmt bool
	)
	for rest := data[12:]; len(rest) >= 8; {
		id, size := string(rest[0:4]), int(binary.LittleEndian.Uint32(rest[4:8]))
		body := rest[8:]
		if size > len(body) {
			if id != "data" {
				return wav{}, fmt.Errorf("chunk '%s' is truncated", id)
			}
			size = len(body)
		}
		body = body[:size]

		switch id {
		case "fmt ":
			if size < 16 {
				return wav{}, fmt.Errorf("fmt chunk is %d bytes; expected at least 16", size)
			}
			format := binary.LittleEndian.Uint16(body[0:2])
			channels := binary.LittleEndian.Uint16(body[2:4])
			w.sampleRate = int(binary.LittleEndian.Uint32(body[4:8]))
			bits := binary.LittleEndian.Uint16(body[14:16])
			if format != 1 || channels != 1 || bits != 16 || w.sampleRate <= 0 {
				return wav{}, fmt.Errorf("format %d, %d channels, %d bits, %d Hz; expected PCM (1), 1 channel, 16 bits",
					format, channels, bits, w.sampleRate)
			}
			hasFmt = true
		case "data":
			if !hasFmt {
				return wav{}, errors.New("data chunk precedes the fmt chunk")
			}
			w.samples = body
			return w, nil
		}

		// Chunks are padded to an even size.
		rest = rest[min(8+size+size%2, len(rest)):]
	}
	return wav{}, errors.New("no data chunk")
}
