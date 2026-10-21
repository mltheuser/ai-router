import type { ModelFilters, ModelList } from "../models";
import type { CallOptions, Transport } from "../transport";
import type { EmbeddingModel, EmbeddingRequest, EmbeddingResponse } from "./embedding";

const USE_CASE = "embedding";

/** The embedding use case: texts in, one vector per text out. */
export class EmbeddingClient {
  readonly #transport: Transport;

  constructor(transport: Transport) {
    this.#transport = transport;
  }

  send(request: EmbeddingRequest, options?: CallOptions): Promise<EmbeddingResponse> {
    return this.#transport.send(USE_CASE, request, options);
  }

  /** The models {@link send} accepts. */
  listModels(filters?: ModelFilters, options?: CallOptions): Promise<ModelList<EmbeddingModel>> {
    return this.#transport.listModels(USE_CASE, filters, options);
  }
}
