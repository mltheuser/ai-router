import type { ModelRef } from "../models";

/** An optional ability of a chat model beyond plain text chat. */
export type ChatFeature = "tools" | "vision" | "reasoning" | "structured_output";

/**
 * A chat model at one provider, as listed by `GET /v1/chat/models`. Cost
 * fields are per million tokens; absent means unknown and `0` means free.
 */
export interface ChatModel extends ModelRef {
  features?: ChatFeature[];
  context_window?: number;
  cost_per_m_input?: number;
  cost_per_m_output?: number;
  /** Set for local models. */
  size_bytes?: number;
}
