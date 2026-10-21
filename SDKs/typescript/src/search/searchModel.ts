import type { ModelRef } from "../models";

/**
 * A search model at one provider, as listed by `GET /v1/search/models`: one
 * way the provider can search, e.g. a fast mode or a deep research mode.
 */
export interface SearchModel extends ModelRef {}
