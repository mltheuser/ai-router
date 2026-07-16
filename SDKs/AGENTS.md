# SDK Conventions

Every SDK in this directory follows the same shape. New SDKs mirror it.
See [kotlin/](kotlin/) for the canonical implementation.

## Examples are tests

Each SDK ships one example file per documented usage pattern, living in the
language's standard test source path (Kotlin: `src/test/...`, Python:
`tests/`, etc.). Each file is one test class/module with one test function,
discovered and run by the language's standard test framework — no custom
runner, no central registry.

Examples serve two roles at once:

- **Documentation.** The SDK's README contains no inline code snippets;
  every section is prose plus a link to the corresponding example file.
- **Drift detection.** Each example runs against a live ai-router server
  (started by the user, not the test) and asserts only that the response
  parsed without error. Don't assert on model output.

Some deviation allowed: for example for error-handling examples that intentionally
trigger a failure and assert on the SDK's exception/result shape.

## Shared setup file

In the same directory as the examples, one file holds:

- Constants `SERVER_URL`, `CHAT_MODEL`, `EMBED_MODEL`. Pick a `CHAT_MODEL`
  that supports vision, reasoning, tools, and structured output — every
  chat-shaped example shares it. Add more constants only on a real
  capability gap.
- A client factory used by every example, configured with a custom HTTP
  layer (timeouts, etc.). This exercises the "custom HTTP client"
  configuration path on every test run instead of needing a separate test.

## Test resources

Binary fixtures (e.g. an image for the vision example) live in the SDK's
own standard test-resources location. Duplicating a small fixture across SDKs is fine.

## Running

Each SDK documents its run command in the README's `## Testing` section
(`./gradlew test`, `pytest`, `cargo test`, …). There is no top-level
orchestrator: SDKs are tested individually. Because examples exercise live
models, a run can occasionally flake on model nondeterminism (e.g. an empty
answer) — re-run the failing example before suspecting the SDK.

Each SDK also runs strict, locally-bootstrapped lint/static-analysis tooling
and is kept at zero findings; the README's `## Linting` section documents the
command and config.
