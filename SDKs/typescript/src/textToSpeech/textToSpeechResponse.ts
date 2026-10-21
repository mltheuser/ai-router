/**
 * The answer to a text-to-speech request. `audio` is a base64-encoded WAV
 * file of 16-bit mono PCM, at a sample rate up to the provider.
 */
export interface TextToSpeechResponse {
  model: string;
  audio: string;
}
