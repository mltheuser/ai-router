package speechtotext

// Request is the body of POST /v1/speech-to-text.
type Request struct {
	Model string `json:"model"`
	// Audio is the speech as a base64-encoded WAV file of 16-bit mono PCM, the
	// format text-to-speech returns. Any sample rate will do; the WAV header
	// carries it.
	Audio string `json:"audio"`
}

// Response is the body of a successful POST /v1/speech-to-text response.
type Response struct {
	Model string `json:"model"`
	// Text is the transcript of the whole audio.
	Text string `json:"text"`
}
