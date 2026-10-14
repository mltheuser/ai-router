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

/** The body of `POST /v1/embedding`: the texts to embed. */
export interface EmbeddingRequest {
  model: string;
  input: string[];
  /** Asks for vectors of this size. Not every model supports it. */
  dimensions?: number;
}

/** The answer to an embedding request: one {@link Embedding} per input text. */
export interface EmbeddingResponse {
  model: string;
  data: Embedding[];
  usage: EmbeddingUsage;
}

/** The vector of the input text at `index`. */
export interface Embedding {
  index: number;
  embedding: number[];
}

/** The tokens a request consumed. */
export interface EmbeddingUsage {
  prompt_tokens: number;
  total_tokens: number;
}
