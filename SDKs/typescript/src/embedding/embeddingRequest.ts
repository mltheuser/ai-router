/** The body of `POST /v1/embedding`: the texts to embed. */
export interface EmbeddingRequest {
  model: string;
  input: string[];
  /** Asks for vectors of this size. Not every model supports it. */
  dimensions?: number;
}
