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
