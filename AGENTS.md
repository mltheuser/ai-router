# AI Router — Project Guide

## Project Overview

`ai-router` is a proxy server for LLM calls that routes requests to various providers (cloud or local) based on configuration and heuristics.

### Core Philosophy
1.  **Unified API**: The server exposes a single, purpose-built API. Clients don't need to know if they are talking to a cloud provider or a local Ollama instance.
2.  **Provider Independence**: Each provider is implemented in isolation within `providers/`, sharing no code beyond the common `Provider` interface — defined in `provider/provider.go`.
3.  **Dynamic Routing**: Requests specify a model using the format `model_id:tag[@provider]`. The router resolves this to the best available provider.
4.  **Tagging System**:
    - `:cloud` - Routes to a cloud provider (e.g., OpenRouter, OpenAI). Selection heuristic: Lowest cost.
    - `:local` - Routes to a local runner (e.g., Ollama). Selection heuristic: First available.
    - `@provider` - Optional suffix to force a specific provider (e.g., `@openrouter`).

## Key Directories

-   **`api/`**: Shared, provider-independent request/response types and error helpers, imported by `router`, `server`, and `providers`.
-   **`router/`**: Core routing — resolves model strings to providers and maintains the model/provider catalog.
-   **`provider/`**: The `Provider` interface and the registry machinery that maps provider names to factories.
-   **`providers/`**: Self-contained provider implementations, one per subdirectory (e.g. `ollama/`, `openrouter/`).
-   **`server/`**: HTTP server and request handlers.
-   **`scenarios/`**: Modular E2E test scenarios that verify provider capabilities.
-   **`cmd/`** / **`cli/`**: Process entry points and CLI commands (e.g. `serve`).
-   **[`SDKs/`](SDKs/)**: Client libraries for the proxy, one per language. Carries its own guide with the conventions every SDK follows — read when working on any SDK.

## Architecture Highlights

### Adding a New Provider
1.  Create a new directory in `providers/<provider_name>`.
2.  Implement the `Provider` interface.
3.  **Register the provider**: Edit `router/registry.go` and add a registration call in `DefaultRegistry()`.

### Verification
Lint and test before finishing a change: `make lint` / `make fmt`, `make test`, and the live scenario runner (`POST /v1/test`). See [TESTING.md](TESTING.md) — read when linting, running the tests, or running the server locally.

## Maintenance for Agents

Update this file when you change the project's **foundations** (the `Provider` interface, routing logic, or testing strictures) — not when adding or modifying individual providers in `providers/`. This file documents the *system*, not its content.