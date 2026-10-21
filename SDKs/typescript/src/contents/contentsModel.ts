import type { ModelRef } from "../models";

/**
 * A contents model at one provider, as listed by `GET /v1/contents/models`:
 * one way the provider can load pages, e.g. from its cache or always fresh.
 */
export interface ContentsModel extends ModelRef {}
