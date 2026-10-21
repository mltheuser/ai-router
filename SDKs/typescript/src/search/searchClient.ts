import type { ModelFilters, ModelList } from "../models";
import type { CallOptions, Transport } from "../transport";
import type { SearchModel } from "./searchModel";
import type { SearchRequest } from "./searchRequest";
import type { SearchResponse } from "./searchResponse";

const USE_CASE = "search";

/** The search use case: a query in, a ranked list of web pages out. */
export class SearchClient {
  readonly #transport: Transport;

  constructor(transport: Transport) {
    this.#transport = transport;
  }

  send(request: SearchRequest, options?: CallOptions): Promise<SearchResponse> {
    return this.#transport.send(USE_CASE, request, options);
  }

  /** The models {@link send} accepts. */
  listModels(filters?: ModelFilters, options?: CallOptions): Promise<ModelList<SearchModel>> {
    return this.#transport.listModels(USE_CASE, filters, options);
  }
}
