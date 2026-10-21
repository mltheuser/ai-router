import type { ModelRef } from "../models";

/**
 * An embedding model at one provider, as listed by `GET /v1/embedding/models`.
 * The cost is per million input tokens; absent means unknown and `0` means
 * free.
 */
export interface EmbeddingModel extends ModelRef {
  context_window?: number;
  cost_per_m_input?: number;
  /** Set for local models. */
  size_bytes?: number;
}
