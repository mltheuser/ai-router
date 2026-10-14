import type { ModelFilters, ModelList } from "../models";
import type { Transport } from "../transport";
import type { TextToSpeechModel, TextToSpeechRequest, TextToSpeechResponse } from "./textToSpeech";

const USE_CASE = "text-to-speech";

/**
 * The text-to-speech use case: a text and a voice in, the text spoken in that
 * voice out, as a WAV file.
 */
export class TextToSpeechClient {
  readonly #transport: Transport;

  constructor(transport: Transport) {
    this.#transport = transport;
  }

  send(request: TextToSpeechRequest): Promise<TextToSpeechResponse> {
    return this.#transport.send(USE_CASE, request);
  }

  /** The models {@link send} accepts, each with its voices. */
  listModels(filters?: ModelFilters): Promise<ModelList<TextToSpeechModel>> {
    return this.#transport.listModels(USE_CASE, filters);
  }
}
