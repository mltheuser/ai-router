# ai-router Kotlin SDK

Kotlin SDK for the [ai-router](../../) LLM proxy. Build requests with a
type-safe DSL, send them to a local or remote proxy, and consume responses —
including structured output via `@Serializable` classes.

Every documented usage pattern is backed by a runnable example in
[src/test/kotlin/ai/router/sdk/examples/](src/test/kotlin/ai/router/sdk/examples/).

The examples double as the SDK's integration test suite — see
[Testing](#testing) below.

## Installation

Add the SDK to your `build.gradle.kts`:

```kotlin
dependencies {
    implementation("ai.router:ai-router-sdk:0.1.0")
}
```

> For now, build from source: `./gradlew publishToMavenLocal`

## Usage

### Quick Start

A minimal end-to-end usage of the SDK: connect to the proxy, send one chat
request, read the text response.

See [QuickStartExample.kt](src/test/kotlin/ai/router/sdk/examples/QuickStartExample.kt).

> Append `@<provider>` to a model string to pin a request to a specific
> backend (e.g. `llama3.2:cloud@openrouter`).

### Multi-Turn Conversation

Send a sequence of system / user / assistant messages and collect a final
response. Demonstrates request-level sampling parameters
(`temperature`, `maxTokens`, …).

See [MultiTurnExample.kt](src/test/kotlin/ai/router/sdk/examples/MultiTurnExample.kt).

### Vision (Multimodal)

Attach an image alongside a text prompt. The image is passed inline as
base64-encoded bytes with an explicit MIME type.

See [VisionExample.kt](src/test/kotlin/ai/router/sdk/examples/VisionExample.kt).

### Tool Calling

Define a tool whose parameters are a `@Serializable` class — the SDK derives
the JSON Schema automatically. The example walks through the full round-trip:
issue the call, decode the model's tool call into the typed class, and send
the tool result back as a follow-up turn. Annotate fields with
[`@Description`](src/main/kotlin/ai/router/sdk/schema/Description.kt) to give
the model human-readable hints.

See [ToolCallingExample.kt](src/test/kotlin/ai/router/sdk/examples/ToolCallingExample.kt).

### Structured Output

Declare the response shape as a `@Serializable` class and use
`structuredChatRequest`. The SDK generates the schema, sets the
`response_format`, and deserializes the response into the typed class — no
manual JSON wrangling.

See [StructuredOutputExample.kt](src/test/kotlin/ai/router/sdk/examples/StructuredOutputExample.kt).

### Reasoning

Request a reasoning-capable model with a target effort level and read both
the visible answer and the model's reasoning trace separately
(via `response.choices.message.reasoningContent`).

See [ReasoningExample.kt](src/test/kotlin/ai/router/sdk/examples/ReasoningExample.kt).

### Embeddings

Each `batch(...)` call corresponds to one embedding in the response. The
optional `dimensions(...)` clause requests a specific embedding size from the
provider.

See [EmbeddingsExample.kt](src/test/kotlin/ai/router/sdk/examples/EmbeddingsExample.kt).

### Error Handling

Non-2xx responses surface as `AiRouterException`, which carries the HTTP
status code and a message of the form `"<error_type>: <error_message>"`.

See [ErrorHandlingExample.kt](src/test/kotlin/ai/router/sdk/examples/ErrorHandlingExample.kt).

### Configuration

Pass a pre-built Ktor `HttpClient` to `AiRouterClient(...)` to take full
control of timeouts, logging, engine choice, etc. The client must have JSON
content negotiation installed. The example tests in this repo all build
their client through a shared factory that demonstrates this pattern — see
[ExampleSetup.kt](src/test/kotlin/ai/router/sdk/examples/ExampleSetup.kt).

## Testing

The example files in
[src/test/kotlin/ai/router/sdk/examples/](src/test/kotlin/ai/router/sdk/examples/)
are runnable JUnit tests that exercise every documented usage pattern against
a live ai-router server. They serve double duty as documentation and as a
drift-detection suite for changes to the server or the SDK.

Prerequisites:

- A running ai-router server. See the project's [testing guide](../../TESTING.md)
  for how to build and run one.
- A provider with the models referenced in
  [ExampleSetup.kt](src/test/kotlin/ai/router/sdk/examples/ExampleSetup.kt).
  Adjust the constants as needed.

Run all examples:

```bash
./gradlew test
```

## Linting

Static analysis runs through [detekt](https://detekt.dev/) (with ktlint
formatting rules folded in via `detekt-formatting`) over both main and test
sources. The build also enforces explicit API mode (strict) and treats all
compiler warnings as errors, so the public API surface stays intentional.

```bash
./gradlew detekt        # report (also runs as part of `./gradlew build`)
./gradlew detekt --auto-correct   # apply the auto-fixable formatting fixes
```

Config lives in [detekt.yml](detekt.yml); it builds on detekt's defaults and
documents every relaxed rule inline.
