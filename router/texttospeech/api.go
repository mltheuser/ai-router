package texttospeech

// Request is the body of POST /v1/text-to-speech.
type Request struct {
	Model string `json:"model"`
	// Voice is the ID of one of the voices listed in the model's entry.
	Voice string `json:"voice"`
	Text  string `json:"text"`
}

// Response is the body of a successful POST /v1/text-to-speech response.
type Response struct {
	Model string `json:"model"`
	// Audio is the speech as a base64-encoded WAV file of 16-bit mono PCM. The
	// sample rate is up to the provider; the WAV header carries it.
	Audio string `json:"audio"`
}
