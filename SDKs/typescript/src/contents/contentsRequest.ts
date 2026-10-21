/** The body of `POST /v1/contents`: the pages to load. */
export interface ContentsRequest {
  model: string;
  urls: string[];
}
