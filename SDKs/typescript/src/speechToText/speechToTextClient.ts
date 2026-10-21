import type { ModelFilters, ModelList } from "../models";
import type { CallOptions, Transport } from "../transport";
import type { SpeechToTextModel, SpeechToTextRequest, SpeechToTextResponse } from "./speechToText";

const USE_CASE = "speech-to-text";

/** The speech-to-text use case: speech in, as a WAV file, its transcript out. */
export class SpeechToTextClient {
  readonly #transport: Transport;

  constructor(transport: Transport) {
    this.#transport = transport;
  }

  /** Audio that is not a WAV file of 16-bit mono PCM is rejected with a 400. */
  send(request: SpeechToTextRequest, options?: CallOptions): Promise<SpeechToTextResponse> {
    return this.#transport.send(USE_CASE, request, options);
  }

  /** The models {@link send} accepts. */
  listModels(filters?: ModelFilters, options?: CallOptions): Promise<ModelList<SpeechToTextModel>> {
    return this.#transport.listModels(USE_CASE, filters, options);
  }
}
