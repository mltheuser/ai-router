import type { ModelFilters, ModelList } from "../models";
import type { Transport } from "../transport";
import type { SearchModel, SearchRequest, SearchResponse } from "./search";

const USE_CASE = "search";

/** The search use case: a query in, a ranked list of web pages out. */
export class SearchClient {
  readonly #transport: Transport;

  constructor(transport: Transport) {
    this.#transport = transport;
  }

  send(request: SearchRequest): Promise<SearchResponse> {
    return this.#transport.send(USE_CASE, request);
  }

  /** The models {@link send} accepts. */
  listModels(filters?: ModelFilters): Promise<ModelList<SearchModel>> {
    return this.#transport.listModels(USE_CASE, filters);
  }
}
