import type { ModelFilters, ModelList } from "../models";
import type { Transport } from "../transport";
import type { ContentsModel, ContentsRequest, ContentsResponse } from "./contents";

const USE_CASE = "contents";

/** The contents use case: web page URLs in, each page's content out. */
export class ContentsClient {
  readonly #transport: Transport;

  constructor(transport: Transport) {
    this.#transport = transport;
  }

  /** A page that fails to load carries an error in its result instead of failing the request. */
  send(request: ContentsRequest): Promise<ContentsResponse> {
    return this.#transport.send(USE_CASE, request);
  }

  /** The models {@link send} accepts. */
  listModels(filters?: ModelFilters): Promise<ModelList<ContentsModel>> {
    return this.#transport.listModels(USE_CASE, filters);
  }
}
