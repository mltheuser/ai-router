import type { ModelRef } from "../models";

/**
 * The bytes per `String.fromCharCode` call: spreading them passes each as an
 * argument, and engines cap the argument count. 32 Ki is well below every cap.
 */
const CHUNK_BYTES = 32_768;

/** A speech-to-text model at one provider, as listed by `GET /v1/speech-to-text/models`. */
export interface SpeechToTextModel extends ModelRef {
  /** The price per hour of input audio; absent means unknown and `0` means free. */
  cost_per_audio_hour?: number;
}

/**
 * The body of `POST /v1/speech-to-text`. `audio` is a base64-encoded WAV file
 * of 16-bit mono PCM, at any sample rate: the format text-to-speech returns,
 * so a text-to-speech response's `audio` can be passed as is.
 * {@link wavBase64} encodes the bytes of a WAV file.
 */
export interface SpeechToTextRequest {
  model: string;
  audio: string;
}

/** The answer to a speech-to-text request: `text` is the transcript of the whole audio. */
export interface SpeechToTextResponse {
  model: string;
  text: string;
}

/**
 * The bytes of a WAV file as a request's `audio`. Portable, if not the
 * fastest: on Node, `Buffer.from(wav).toString("base64")` gives the same
 * string.
 */
export function wavBase64(wav: Uint8Array): string {
  let binary = "";
  for (let start = 0; start < wav.length; start += CHUNK_BYTES) {
    binary += String.fromCharCode(...wav.subarray(start, start + CHUNK_BYTES));
  }
  return btoa(binary);
}
