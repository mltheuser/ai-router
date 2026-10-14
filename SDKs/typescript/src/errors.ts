/** The error body ai-router answers a failed request with. */
export interface ApiError {
  type: string;
  message: string;
  code?: string;
}

/** The wrapper around {@link ApiError} on the wire. */
export interface ErrorResponse {
  error: ApiError;
}

/**
 * Thrown when ai-router answers with a non-2xx status. A body that is not an
 * {@link ErrorResponse} becomes an error of type `unknown` with the raw body
 * as its message.
 */
export class AiRouterError extends Error {
  override readonly name = "AiRouterError";
  readonly status: number;
  readonly apiError: ApiError;

  constructor(status: number, apiError: ApiError) {
    super(`${apiError.type}: ${apiError.message}`);
    this.status = status;
    this.apiError = apiError;
  }
}
