import type { ModelFilters, ModelList } from "../models";
import type { CallOptions, Transport } from "../transport";
import type { ContentsModel } from "./contentsModel";
import type { ContentsRequest } from "./contentsRequest";
import type { ContentsResponse } from "./contentsResponse";

const USE_CASE = "contents";

/** The contents use case: web page URLs in, each page's content out. */
export class ContentsClient {
  readonly #transport: Transport;

  constructor(transport: Transport) {
    this.#transport = transport;
  }

  /** A page that fails to load carries an error in its result instead of failing the request. */
  send(request: ContentsRequest, options?: CallOptions): Promise<ContentsResponse> {
    return this.#transport.send(USE_CASE, request, options);
  }

  /** The models {@link send} accepts. */
  listModels(filters?: ModelFilters, options?: CallOptions): Promise<ModelList<ContentsModel>> {
    return this.#transport.listModels(USE_CASE, filters, options);
  }
}
