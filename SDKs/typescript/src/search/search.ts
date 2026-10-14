import type { ModelRef } from "../models";

/**
 * A search model at one provider, as listed by `GET /v1/search/models`: one
 * way the provider can search, e.g. a fast mode or a deep research mode.
 */
export interface SearchModel extends ModelRef {}

/** The body of `POST /v1/search`. */
export interface SearchRequest {
  model: string;
  query: string;
  /** Caps the number of results; the provider may return fewer. */
  max_results?: number;
}

/** The answer to a search request: web pages, most relevant first. */
export interface SearchResponse {
  model: string;
  results: SearchResult[];
}

/** One web page found for the query. */
export interface SearchResult {
  url: string;
  /** May be empty: not every page has a title. */
  title: string;
  /** The excerpts the provider selected as relevant to the query; not necessarily short. */
  snippet: string;
}
