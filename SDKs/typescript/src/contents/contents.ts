import type { ModelRef } from "../models";

/**
 * A contents model at one provider, as listed by `GET /v1/contents/models`:
 * one way the provider can load pages, e.g. from its cache or always fresh.
 */
export interface ContentsModel extends ModelRef {}

/** The body of `POST /v1/contents`: the pages to load. */
export interface ContentsRequest {
  model: string;
  urls: string[];
}

/** The answer to a contents request: one {@link ContentsResult} per requested URL, in request order. */
export interface ContentsResponse {
  model: string;
  results: ContentsResult[];
}

/**
 * The content of one requested page. A page that could not be loaded has
 * `error` set, the provider's reason, and no `title` or `text`; the request
 * as a whole still succeeds.
 */
export interface ContentsResult {
  /** The URL as requested. */
  url: string;
  title?: string;
  /**
   * The page's content, in a format up to the provider. A provider that keeps
   * the page's images includes them inline, as links.
   */
  text?: string;
  error?: string;
}
