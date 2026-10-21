import { AiRouterError, type ApiError, type ErrorResponse } from "./errors";
import type { ModelFilters, ModelList, ModelRef } from "./models";

/** LLM calls can run long; allow up to 10 minutes per request. */
const REQUEST_TIMEOUT_MILLIS = 600_000;

/** Reads a 2xx response's body as `T`; any other status throws {@link AiRouterError}. */
async function decode<T>(response: Response): Promise<T> {
  if (!response.ok) {
    throw new AiRouterError(response.status, parseApiError(await response.text()));
  }
  return (await response.json()) as T;
}

function parseApiError(body: string): ApiError {
  try {
    const error = (JSON.parse(body) as Partial<ErrorResponse> | null)?.error;
    if (typeof error?.type === "string" && typeof error.message === "string") {
      return error;
    }
  } catch {
    // Not JSON: fall through to the raw body.
  }
  return { type: "unknown", message: body };
}

function signalOf(options: CallOptions): AbortSignal {
  const timeout = AbortSignal.timeout(REQUEST_TIMEOUT_MILLIS);
  if (options.signal === undefined) {
    return timeout;
  }
  return AbortSignal.any([options.signal, timeout]);
}

/** What every SDK call accepts besides its request. */
export interface CallOptions {
  /** Cancels the call when it aborts, e.g. `AbortSignal.timeout(30_000)`. */
  signal?: AbortSignal;
}

export class Transport {
  readonly #baseUrl: string;

  constructor(baseUrl: string) {
    this.#baseUrl = baseUrl;
  }

  async send<Res>(useCase: string, request: unknown, options: CallOptions = {}): Promise<Res> {
    const response = await fetch(`${this.#baseUrl}/v1/${useCase}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(request),
      signal: signalOf(options),
    });
    return decode(response);
  }

  async listModels<M extends ModelRef>(
    useCase: string,
    filters: ModelFilters = {},
    options: CallOptions = {},
  ): Promise<ModelList<M>> {
    const url = new URL(`${this.#baseUrl}/v1/${useCase}/models`);
    if (filters.type !== undefined) {
      url.searchParams.set("type", filters.type);
    }
    if (filters.search !== undefined) {
      url.searchParams.set("search", filters.search);
    }
    const response = await fetch(url.href, { method: "GET", signal: signalOf(options) });
    return decode(response);
  }
}
