import { AiRouterError, type ApiError, type ErrorResponse } from "./errors";
import type { ModelFilters, ModelList, ModelRef } from "./models";

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

/** The subset of `fetch` the SDK calls: always a URL string and an init. */
export type Fetch = (url: string, init: RequestInit) => Promise<Response>;

/**
 * The two routes every use case is served on, `POST /v1/<use case>` and
 * `GET /v1/<use case>/models`, over one `fetch`.
 */
export class Transport {
  readonly #baseUrl: string;
  readonly #fetch: Fetch;

  constructor(baseUrl: string, fetch: Fetch) {
    this.#baseUrl = baseUrl;
    // Wrapped so that fetch is never called with this transport as `this`,
    // which browsers reject for the global fetch.
    this.#fetch = (url, init) => fetch(url, init);
  }

  async send<Res>(useCase: string, request: unknown): Promise<Res> {
    const response = await this.#fetch(`${this.#baseUrl}/v1/${useCase}`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(request),
    });
    return decode(response);
  }

  async listModels<M extends ModelRef>(
    useCase: string,
    filters: ModelFilters = {},
  ): Promise<ModelList<M>> {
    const url = new URL(`${this.#baseUrl}/v1/${useCase}/models`);
    if (filters.type !== undefined) {
      url.searchParams.set("type", filters.type);
    }
    if (filters.search !== undefined) {
      url.searchParams.set("search", filters.search);
    }
    const response = await this.#fetch(url.href, { method: "GET" });
    return decode(response);
  }
}
