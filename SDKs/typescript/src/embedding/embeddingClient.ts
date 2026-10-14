import type { ModelFilters, ModelList } from "../models";
import type { Transport } from "../transport";
import type { EmbeddingModel, EmbeddingRequest, EmbeddingResponse } from "./embedding";

const USE_CASE = "embedding";

/** The embedding use case: texts in, one vector per text out. */
export class EmbeddingClient {
  readonly #transport: Transport;

  constructor(transport: Transport) {
    this.#transport = transport;
  }

  send(request: EmbeddingRequest): Promise<EmbeddingResponse> {
    return this.#transport.send(USE_CASE, request);
  }

  /** The models {@link send} accepts. */
  listModels(filters?: ModelFilters): Promise<ModelList<EmbeddingModel>> {
    return this.#transport.listModels(USE_CASE, filters);
  }
}
