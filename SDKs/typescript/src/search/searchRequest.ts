/** The body of `POST /v1/search`. */
export interface SearchRequest {
  model: string;
  query: string;
  /** Caps the number of results; the provider may return fewer. */
  max_results?: number;
}
