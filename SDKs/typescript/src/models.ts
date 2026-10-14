/** Distinguishes cloud from local providers. */
export type ProviderType = "cloud" | "local";

/**
 * The routing identity every listed model carries, whatever its use case.
 * `model` is the fully-qualified string (`id:provider_type@provider`) to pass
 * verbatim as a request's `model`.
 */
export interface ModelRef {
  id: string;
  model: string;
  provider: string;
  provider_type: ProviderType;
}

/** The body of a use case's model listing, e.g. `GET /v1/chat/models`. */
export interface ModelList<M extends ModelRef> {
  data: M[];
}

/**
 * Narrows a model listing by provider type and a case-insensitive substring
 * of the model id.
 */
export interface ModelFilters {
  type?: ProviderType;
  search?: string;
}
