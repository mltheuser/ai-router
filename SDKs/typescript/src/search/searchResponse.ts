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
