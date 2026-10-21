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
