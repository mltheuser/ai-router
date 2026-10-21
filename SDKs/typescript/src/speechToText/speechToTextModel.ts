import type { ModelRef } from "../models";

/** A speech-to-text model at one provider, as listed by `GET /v1/speech-to-text/models`. */
export interface SpeechToTextModel extends ModelRef {
  /** The price per hour of input audio; absent means unknown and `0` means free. */
  cost_per_audio_hour?: number;
}
