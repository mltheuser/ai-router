# SDK Conventions

Every SDK in this directory follows the same shape. New SDKs mirror it.
See [kotlin/](kotlin/) for the canonical implementation.

## API surface mirrors the use cases

The server is organized by use case (e.g. chat, embedding), and so is every
SDK. Each use case gets its own module holding its
request/response types, a model type with the use case's own metadata, and a
use-case client with two methods: `send`, the use case's request, and
`listModels`, returning that model type. Types are named after the use case
(`EmbeddingRequest`, `EmbeddingResponse`, `EmbeddingModel`). The model types
share one interface for the routing fields (`id`, `model`, `provider`,
`provider_type`).

Types follow the wire format.

The top-level client only wires things together: it exposes one use-case
client per use case as a property named after the use case
(`client.chat.send(...)`), and one internal transport that serves them all
(`POST /v1/<use case>`, `GET /v1/<use case>/models`, error decoding). No class
grows with the number of use cases beyond one property on the top-level
client.

The top-level client takes the server's base URL and nothing else: the SDK
owns its HTTP setup (timeouts, JSON handling), so a caller cannot configure it
into a state the SDK does not work in.

Adding a use case to the server means adding a module to each SDK and one
property to its top-level client.

## Examples are tests

Each SDK ships one example file per documented usage pattern, living in the
language's standard test source path. Each file is one test class/module with one
test function, discovered and run by the language's standard test framework —
no custom runner, no central registry.

Examples serve two roles at once:

- **Documentation.** The SDK's README contains no inline code snippets;
  every section is prose plus a link to the corresponding example file.
- **Drift detection.** Each example runs against a live ai-router server
  (started by the user, not the test) and asserts only that the response
  parsed without error. Don't assert on model output.

Some deviation allowed: for example for error-handling examples that intentionally
trigger a failure and assert on the SDK's exception/result shape.

## Shared setup file

In the same directory as the examples, one file holds `SERVER_URL`, and one
model constant per use case (`CHAT_MODEL`, `EMBED_MODEL`, ...). Pick a
`CHAT_MODEL` with every chat feature (vision, reasoning, ...) — every example shares it. Add more constants only on a real
capability gap.

## Test resources

Binary fixtures (e.g. an image for the vision example) live in the SDK's own
standard test-resources location. Duplicating a small fixture across SDKs is fine.

## Running

Each SDK documents its run command in the README's `## Testing` section. There is no
top-level orchestrator: SDKs are tested individually. Because examples
exercise live models, a run can occasionally flake on model nondeterminism
(e.g. an empty answer) — re-run the failing example before suspecting the SDK.

Each SDK also runs strict, locally-bootstrapped lint/static-analysis tooling
and is kept at zero findings; the README's `## Linting` section documents the
command and config.
