# AI Router — Project Guide

## Project Overview

`ai-router` is a proxy server for LLM calls (OpenAI-compatible) that routes requests to various providers (cloud or local) based on configuration and heuristics.

### Core Philosophy
1.  **Unified API**: The server exposes a single OpenAI-compatible API. Clients don't need to know if they are talking to OpenAI, Anthropic, or a local Ollama instance.
2.  **Provider Independence**: Each provider is implemented in isolation within `providers/`. They share no code other than the common interface defined in `provider/`.
3.  **Dynamic Routing**: Requests specify a model using the format `model_id:tag[@provider]`. The router resolves this to the best available provider.
4.  **Tagging System**:
    - `:cloud` - Routes to a cloud provider (e.g., OpenRouter, OpenAI). Selection heuristic: Lowest cost.
    - `:local` - Routes to a local runner (e.g., Ollama). Selection heuristic: First available.
    - `@provider` - Optional suffix to force a specific provider (e.g., `@openrouter`).

## Key Directories

-   **`api/`**: Shared API types and errors.
    -   `types.go`, `errors.go`: Common definitions imported by `router`, `server`, and `providers`.
-   **`router/`**: Core routing logic.
    -   `router.go`: Resolves model strings to providers.
    -   `catalog.go`: Maintains the list of available models and providers.
-   **`provider/`**: Provider interface and registry machinery.
    -   `provider.go`: Defines the `Provider` interface.
    -   `registry.go`: The `Registry` type mapping provider names to factories (avoids switch statements). The default registry that wires up known providers lives in `router/registry.go` — see [Adding a New Provider](#adding-a-new-provider).
-   **`providers/`**: Implementations.
    -   Each subdirectory (e.g., `ollama/`, `openrouter/`) contains a self-contained implementation.
-   **`server/`**: HTTP Server and Handlers.
    -   `server.go`: Main server entry point. Registers handlers for `/v1/models`, `/v1/embeddings`, `/v1/chat/completions`, `/v1/test`, and `/health`.
    -   `handle_run_tests.go`: Implementation of the `/v1/test` endpoint.
-   **`scenarios/`**: E2E Test Scenarios.
    -   Contains modular test scenarios that verify provider capabilities (e.g., `EmbedBatchSimilarity`).
-   **`cmd/`**: Entry points.
-   **`cli/`**: CLI command implementations (e.g., `serve`).
-   **[`SDKs/`](SDKs/)**: Client libraries for the proxy, one per language. Carries its own guide with the conventions every SDK follows — read when working on any SDK.

## Architecture Highlights

### The Provider Interface
Every provider implements the `Provider` interface — `Name`, `Type`, `Verify`, `ListModels`, `Embed`, and `Chat`. The interface and its doc comments are the source of truth in `provider/provider.go`.

### Adding a New Provider
1.  Create a new directory in `providers/<provider_name>`.
2.  Implement the `Provider` interface.
3.  **Register the provider**: Edit `router/registry.go` and add a registration call in `DefaultRegistry()`.

### Testing
Two tiers: unit tests (`make test`) for logic, and a built-in scenario runner (`POST /v1/test`) that verifies live providers end to end. See [TESTING.md](TESTING.md) — read when running either suite or running the server locally.

## Maintenance for Agents

-   **Sync Policy**: If you modify the **foundations** of the project (e.g., changing the `Provider` interface, routing logic, or testing strictures), you **MUST** update this document.
-   **No Sync Needed**: You do **NOT** need to update this file when simply adding, removing, or modifying individual providers in `providers/`. This file is about the *system*, not the content.