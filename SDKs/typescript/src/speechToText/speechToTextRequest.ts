/**
 * The body of `POST /v1/speech-to-text`. `audio` is a base64-encoded WAV file
 * of 16-bit mono PCM, at any sample rate: the format text-to-speech returns,
 * so a text-to-speech response's `audio` can be passed as is.
 */
export interface SpeechToTextRequest {
  model: string;
  audio: string;
}
