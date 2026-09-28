# Testing

Project relies on a built-in
scenario runner for end-to-end provider verification. Unit tests are the
exception, and there are no classic integration tests.

During development, write whatever throwaway tests you need, but do not commit
them. What lands in the tree is E2E scenarios; a unit test is only committed when
the owner has explicitly approved it.

## Linting

-   **Commands**: `make lint` (report) and `make fmt` (auto-format). Both bootstrap a
    pinned `golangci-lint` into `./bin` on first run; no global install needed.
-   **Config**: `.golangci.yml` (v2 schema) at the repo root.
-   **Scope**: The whole module (`./...`). `make lint` must exit clean: findings are
    fixed, never suppressed (no `//nolint` directives).

## Unit Tests

-   **Command**: `make test`
-   **Scope**: Isolated units of non-trivial logic with no external dependencies. No mocking. New unit tests
    require owner sign-off before they are committed.
-   **Location**: `*_test.go` files next to the code (e.g. `usecase/usecase_test.go`).

## End-to-End Tests (`/v1/test`)

The primary way to verify providers is the centralized, scenario-based E2E runner built into the server.

1.  **Build**: `make build` (always rebuild after changes).
2.  **Run Server**: `set -a && source .env && set +a && ./bin/ai-router serve --debug`
    - Cloud provider API keys live in `.env` at the project root (not committed); start from `cp .env.example .env`.
    - Key naming convention matches the server's expected format: `AI_ROUTER_<PROVIDER>_API_KEY` (e.g. `AI_ROUTER_OPENROUTER_API_KEY`).
3.  **Trigger**: `curl -X POST http://localhost:8787/v1/test -d '{"provider": "ollama"}' | jq .`
    - Optionally limit the run to one use case: `-d '{"provider": "openrouter", "use_case": "chat"}'`.
    - Optionally pin a model: `-d '{"provider": "openrouter", "use_case": "chat", "model": "~anthropic/claude-sonnet-latest"}'`. The `model` value is the bare `id` as listed by `GET /v1/<use case>/models`. When omitted, each scenario runs against the provider's preferred model it applies to.
    - `serve --addr 127.0.0.1:<port>` runs a second server next to one already on the default port.

**What happens**: the server runs the provider's `Verify()`, then, for every use case the provider serves, lists its models and runs the use case's scenarios through the use case's own endpoint. The report has one section per use case.

**Scenarios**: defined in each use case's package (`usecase/<name>/scenarios.go`). Each scenario may declare which models it applies to (e.g. chat models with the `tools` feature) and is skipped, not failed, when no model qualifies.
