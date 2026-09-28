# AI Router — Project Guide

## Project Overview

`ai-router` is a proxy server that routes AI requests (chat, embedding, ...) to various providers, cloud or local, through one purpose-built API per use case.

### Core Philosophy
1.  **Use cases are the unit of contract.** Each use case (`chat`, `embedding`, ...) owns its API end to end: request and response types, a model type with its own metadata, the provider interface a backend implements to serve it, and its E2E scenarios. Use cases share no types with each other.
2.  **Providers are the unit of connection.** A provider (API key, base URL, `Verify`) is created once and serves every use case whose interface it implements. Serving a use case is opting in by implementing its interface; nothing else needs to know. There are no "not supported" stubs.
3.  **Provider independence.** Each provider lives in isolation in `providers/`, sharing no code with other providers beyond the small helpers in `providers/httpclient`.
4.  **Dynamic routing.** Requests name a model as `model_id:tag[@provider]`. Clients never compose this string themselves: every entry of a use case's model listing carries it in the `model` field, passed verbatim in requests.
    - `:cloud` / `:local` - the provider type. Among providers of that type listing the model, the use case's preference picks one (chat and embedding: cheapest in the cloud, smallest locally).
    - `@provider` - optional suffix to force a specific provider (e.g. `@openrouter`).

## API

Every use case is served the same way:

| Route | Purpose |
|---|---|
| `POST /v1/<use case>` | The request, e.g. `POST /v1/chat`. |
| `GET /v1/<use case>/models` | The models the request accepts, in the use case's own shape. Filters: `?type=cloud\|local`, `?search=<id substring>`. |
| `POST /v1/<use case>/models/refresh` | Re-list the models now instead of waiting for the next refresh. |

Plus `POST /v1/test` (see [TESTING.md](TESTING.md)) and `GET /health`.

## Key Directories

-   **`usecase/`**: The machinery every use case shares, written once and generic over the use case's model and provider types: model catalog and refresh, model-string resolution, the model listing endpoint, and the scenario runner.
-   **`usecase/<name>/`**: One package per use case (e.g. `usecase/chat/`), holding its entire contract and its scenarios.
-   **`provider/`**: What every provider is (`Provider`: name, type, `Verify`) and the routing identity every listed model carries (`ModelRef`).
-   **`providers/`**: Self-contained provider implementations, one per subdirectory (e.g. `ollama/`, `openrouter/`).
-   **`server/`**: The HTTP server. It knows no use case in particular; every route is derived from the `usecase.UseCase` interface.
-   **`api/`**: HTTP plumbing every handler shares: the error type and JSON helpers.
-   **`debug/`**: The `--debug` request log, an HTTP middleware.
-   **`cli/`** / **`cmd/`**: Process entry point and commands. `cli/registry.go` is the one place that knows every provider and every use case.
-   **[`SDKs/`](SDKs/)**: Client libraries for the proxy, one per language. Carries its own guide with the conventions every SDK follows — read when working on any SDK.

## Architecture Highlights

Dependencies point one way: `providers/*` → `usecase/<name>` → `usecase` → `provider`, `api`. A use case never imports a provider, and a provider never imports the server. Only `cli/registry.go` sees both sides.

### Adding a Provider
1.  Create `providers/<name>/` with a type implementing `provider.Provider`.
2.  For each use case it serves, implement that use case's `Provider` interface (e.g. `chat.Provider`: `ListChatModels` and `Chat`), and add a compile-time assertion (`var _ chat.Provider = (*Provider)(nil)`) so a signature mismatch fails the build instead of silently dropping the use case.
3.  Register the constructor in `cli/registry.go`.

### Adding a Use Case
1.  Create `usecase/<name>/` with: a `Model` type embedding `provider.ModelRef`, the `Request`/`Response` types, a `Provider` interface (embedding `provider.Provider`, plus a listing method and a request method with use-case-specific names, since one provider type may implement several use cases), and a `UseCase` type that embeds `*usecase.Base` built from a `usecase.Spec` and adds `Handle`.
2.  Write its scenarios in the same package (`Spec.Scenarios`).
3.  Add it to `useCases` in `cli/registry.go`.
4.  Add the use case to every SDK (see [SDKs/AGENTS.md](SDKs/AGENTS.md)).

`usecase/embedding/` is the smallest complete example.

### Verification
Lint and test before finishing a change: `make lint` / `make fmt`, `make test`, and the live scenario runner (`POST /v1/test`). See [TESTING.md](TESTING.md) — read when linting, running the tests, or running the server locally.

## Maintenance for Agents

Update this file when you change the project's **foundations** (the `provider.Provider` or `usecase.UseCase` contracts, the shared `usecase` machinery, routing, or testing strictures) — not when adding or modifying individual providers or use cases. This file documents the *system*, not its content.
