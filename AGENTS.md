# AI Router — Project Guide

## Project Overview

`ai-router` is a proxy server that routes AI requests (chat, embedding, ...) to various providers, cloud or local, through one purpose-built API per use case.

### Core Philosophy
1.  **Use cases are the unit of contract.** Each use case (`chat`, `embedding`, ...) owns its API end to end: request and response types, a model type with its own metadata, the provider interface a backend implements to serve it, and its E2E scenarios. Use cases share no types with each other.
2.  **Providers are the unit of connection.** A provider (API key, base URL, `Verify`) is created once and serves every use case whose interface it implements. Serving a use case is opting in by implementing its interface; nothing else needs to know. There are no "not supported" stubs.
3.  **Provider independence.** Each provider lives in isolation in `providers/`, sharing no code with other providers; they call their APIs through the shared `httpx.Client`.
4.  **Dynamic routing.** Requests name a model as `model_id:tag[@provider]`. Clients never compose this string themselves: every entry of a use case's model listing carries it in the `model` field, passed verbatim in requests.
    - `:cloud` / `:local` - the provider type. When multiple providers of that type list the model, the use case's heuristic picks the best one.
    - `@provider` - optional suffix to force a specific provider (e.g. `@openrouter`).

## API

Every use case is served the same way:

| Route | Purpose |
|---|---|
| `POST /v1/<use case>` | The request, e.g. `POST /v1/chat`. |
| `GET /v1/<use case>/models` | The models the request accepts, in the use case's own shape. Filters: `?type=cloud\|local`, `?search=<id substring>`. |

Plus `POST /v1/test` (see [TESTING.md](TESTING.md)) and `GET /health`.

Model lists are fetched once, at startup, and never refreshed: restart the server to pick up new models (e.g. a freshly pulled Ollama model).

## Key Directories

The top level separates **what the router defines** (`router/`) from **what plugs into it** (`providers/`); the rest is plumbing and entry points.

-   **`router/`**: Everything the router defines. Its package doc is the map.
    -   `provider.go`: what every provider is (`Provider`: name, type, `Verify`) and the routing identity every listed model carries (`ModelRef`).
    -   `usecase.go`: what the server serves (`UseCase`), and how a use case declares itself (`Spec`) to get the shared machinery (`Base`).
    -   `catalog.go`, `resolve.go`, `test.go`: that machinery, written once and generic over the use case's model and provider types: model catalog and listing endpoint, model-string resolution, and the scenario runner.
-   **`router/<name>/`**: One package per use case (e.g. `router/chat/`), holding its entire contract and its scenarios.
-   **`providers/`**: Self-contained provider implementations, one per subdirectory (e.g. `ollama/`, `openrouter/`).
-   **`server/`**: The HTTP server. It knows no use case in particular; every route is derived from the `router.UseCase` interface.
-   **`httpx/`**: HTTP plumbing shared across the router: the error type every handler responds with, JSON helpers, and the `Client` providers use to call their upstream APIs (which maps upstream failures onto that error type and records exchanges for `--debug`).
-   **`debug/`**: The `--debug` request log, an HTTP middleware.
-   **`cli/`** / **`cmd/`**: Process entry point and commands. `cli/registry.go` is the one place that knows every provider and every use case.
-   **[`SDKs/`](SDKs/)**: Client libraries for the proxy, one per language. Carries its own guide with the conventions every SDK follows — read when working on any SDK.

## Architecture Highlights

Dependencies point one way: `providers/*` → `router/<name>` → `router` → `httpx`. A use case never imports a provider, and a provider never imports the server. Only `cli/registry.go` sees both sides.

### Adding a Provider
1.  Create `providers/<name>/` with a type implementing `router.Provider`.
2.  For each use case it serves, implement that use case's `Provider` interface (e.g. `chat.Provider`: `ListChatModels` and `Chat`), and add a compile-time assertion (`var _ chat.Provider = (*Provider)(nil)`) so a signature mismatch fails the build instead of silently dropping the use case.
3.  Register the constructor in `cli/registry.go`.

### Adding a Use Case
1.  Create `router/<name>/` with three files. `api.go`: the wire types clients send and receive (`Request`, `Response`, ...). `scenarios.go`: the E2E scenarios (`Spec.Scenarios`). `<name>.go`: everything else, namely a `Model` type embedding `router.ModelRef`, a `Provider` interface (embedding `router.Provider`, plus a listing method and a request method with use-case-specific names, since one provider type may implement several use cases), and a `UseCase` type that embeds `*router.Base` built from a `router.Spec` and adds `Handle`.
2.  Add it to `useCases` in `cli/registry.go`.
3.  Add the use case to every SDK (see [SDKs/AGENTS.md](SDKs/AGENTS.md)).

`router/embedding/` is the smallest complete example.

### Verification
Lint and test before finishing a change: `make lint` / `make fmt`, `make test`, and the live scenario runner (`POST /v1/test`). See [TESTING.md](TESTING.md) — read when linting, running the tests, or running the server locally.

## Maintenance for Agents

Update this file when you change the project's **foundations** (the `router.Provider` or `router.UseCase` contracts, the shared machinery in `router/`, routing, or testing strictures) — not when adding or modifying individual providers or use cases. This file documents the *system*, not its content.
