import type { ModelRef } from "../models";

/**
 * A text-to-speech model at one provider, as listed by
 * `GET /v1/text-to-speech/models`, with the voices a request to it may name.
 */
export interface TextToSpeechModel extends ModelRef {
  voices: TextToSpeechVoice[];
  /** The longest text, in characters, the model accepts per request; absent means unknown. */
  max_input_chars?: number;
  /** The price per million input characters; absent means unknown and `0` means free. */
  cost_per_m_chars?: number;
}

/** A voice a model speaks in. Pass its `id` as a request's `voice`. */
export interface TextToSpeechVoice {
  id: string;
  name: string;
}

/** The body of `POST /v1/text-to-speech`: `text` spoken in `voice`, one of the model's voices. */
export interface TextToSpeechRequest {
  model: string;
  voice: string;
  text: string;
}

/**
 * The answer to a text-to-speech request. `audio` is a base64-encoded WAV
 * file of 16-bit mono PCM, at a sample rate up to the provider; {@link wavBytes}
 * decodes it.
 */
export interface TextToSpeechResponse {
  model: string;
  audio: string;
}

/** The response's audio as the bytes of a WAV file. */
export function wavBytes(response: TextToSpeechResponse): Uint8Array {
  return Uint8Array.from(atob(response.audio), (char) => char.charCodeAt(0));
}
