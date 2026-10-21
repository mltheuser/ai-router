/** The body of `POST /v1/text-to-speech`: `text` spoken in `voice`, one of the model's voices. */
export interface TextToSpeechRequest {
  model: string;
  voice: string;
  text: string;
}
