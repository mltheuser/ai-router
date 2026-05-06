# AGENT.md - AI Router Project Guide

This document is designed to help AI coding agents quickly understand the structure, philosophy, and workflow of the `ai-router` project. Use this guide to orient yourself before making changes.

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
-   **`provider/`**: Interfaces and Registry.
    -   `provider.go`: Defines the `Provider` interface.
    -   `registry.go`: Central registry to avoid switch statements. **New providers must be registered here.**
-   **`providers/`**: Implementations.
    -   Each subdirectory (e.g., `ollama/`, `openrouter/`) contains a self-contained implementation.
-   **`server/`**: HTTP Server and Handlers.
    -   `server.go`: Main server entry point. Registers handlers for `/v1/models`, `/v1/embeddings`, `/v1/chat/completions`, `/v1/test`, and `/health`.
    -   `handle_run_tests.go`: Implementation of the `/v1/test` endpoint.
-   **`scenarios/`**: E2E Test Scenarios.
    -   Contains modular test scenarios that verify provider capabilities (e.g., `EmbedBatchSimilarity`).
-   **`cmd/`**: Entry points.
-   **`cli/`**: CLI command implementations (e.g., `serve`).

## Architecture Highlights

### The Provider Interface
Every provider must implement the `Provider` interface:
```go
type Provider interface {
    // Unique identifier (e.g., "ollama")
    Name() string
    
    // Cloud vs Local
    Type() api.ProviderType
    
    // Called on startup and during tests.
    // Cloud: Validate API key. Local: Check if process is running/reachable.
    Verify(ctx context.Context) error
    
    // Fetch available models from the provider.
    ListModels(ctx context.Context) ([]api.ModelInfo, error)
    
    // Generate embeddings.
    Embed(ctx context.Context, req *api.EmbedRequest) (*api.EmbedResponse, error)

    // Generate chat completions.
    Chat(ctx context.Context, req *api.ChatRequest) (*api.ChatResponse, error)
}
```

### Adding a New Provider
1.  Create a new directory in `providers/<provider_name>`.
2.  Implement the `Provider` interface.
3.  **Register the provider**: Edit `router/registry.go` and add a registration call in `DefaultRegistry()`.

### Testing Strategy

The project uses a two-tiered testing approach: unit tests for logic and a built-in scenario runner for end-to-end provider verification.

#### 1. Unit Tests
-   **Command**: `make test`
-   **Scope**: Foundational logic (router resolution, candidate selection) without external dependencies.
-   **Location**: `*_test.go` files next to the code (e.g., `router/router_test.go`).

#### 2. Integration Tests
-   **Status**: There are **NO** classic integration tests run via `make`.
-   **Note**: Do not rely on legacy `make test-integration-*` targets.

#### 3. End-to-End Tests (`/v1/test`)
The primary way to verify providers is the centralized, scenario-based E2E runner built into the server.

-   **Workflow**:
    1.  **Build**: `make build` (Crucial: always rebuild after changes).
    2.  **Run Server**: `set -a && source .env && set +a && ./bin/ai-router serve --debug`
        - Cloud provider API keys live in `.env` at the project root (not committed).
        - Key naming convention matches the server's expected format: `AI_ROUTER_<PROVIDER>_API_KEY` (e.g. `AI_ROUTER_OPENROUTER_API_KEY`).
        - As new providers are added, their test keys are added to `.env` under the same naming convention.
    3.  **Trigger**: `curl -X POST http://localhost:8787/v1/test -d '{"provider": "ollama"}' | jq .`
        - Optionally pin a specific model: `curl -X POST http://localhost:8787/v1/test -d '{"provider": "openrouter", "model": "~anthropic/claude-sonnet-latest"}' | jq .`
        - The `model` value must match the exact ID as returned by `GET /v1/models`. When omitted, the best available model per scenario is auto-selected.
-   **What happens**: The server self-verifies by running `Verify()`, `ListModels()`, and executing applicable scenarios from `scenarios/`.
-   **Scenarios**: Defined in `scenarios/`. Each scenario declares its `RequiredCapabilities()` and runs a specific functional test. Scenarios are skipped (not failed) when the target model lacks a required capability.

## Maintenance for Agents

-   **Sync Policy**: If you modify the **foundations** of the project (e.g., changing the `Provider` interface, routing logic, or testing strictures), you **MUST** update this document.
-   **No Sync Needed**: You do **NOT** need to update this file when simply adding, removing, or modifying individual providers in `providers/`. This file is about the *system*, not the content.