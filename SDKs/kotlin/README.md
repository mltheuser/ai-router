# ai-router Kotlin SDK

Kotlin SDK for the [ai-router](../../) proxy.

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

The client mirrors the server: one property per use case (e.g. `client.chat`,
`client.embedding`) — each with `send(...)`
for the use case's request and `listModels(...)` for the models it accepts.

### Quick Start

A minimal end-to-end usage of the SDK: connect to the proxy, send one chat
request, read the text response.

See [QuickStartExample.kt](src/test/kotlin/ai/router/sdk/examples/QuickStartExample.kt).

> Append `@<provider>` to a model string to pin a request to a specific
> backend (e.g. `llama3.2:cloud@openrouter`).

### Multi-Turn Conversation

Hold a conversation as a `MutableList<ChatMessage>`: add each reply
(`response.message`) as received, then the next user message, and send the
list again. Settings are a plain value, so a conversation may change model or
effort between requests. Demonstrates the sampling parameters (`temperature`,
`maxTokens`, …) of `ChatSettings`.

See [MultiTurnExample.kt](src/test/kotlin/ai/router/sdk/examples/MultiTurnExample.kt).

### Vision (Multimodal)

Attach an image alongside a text prompt: a user message with a text part and
an image part. The image is passed inline as base64-encoded bytes with an
explicit MIME type.

See [VisionExample.kt](src/test/kotlin/ai/router/sdk/examples/VisionExample.kt).

### Tool Calling

Define a tool whose parameters are a `@Serializable` class:
`SchemaGenerator.generate(serializer)` derives its JSON Schema. Annotate fields
with [`@Description`](src/main/kotlin/ai/router/sdk/schema/Description.kt) to
give the model human-readable hints. The example walks through the full
round-trip: add the model's reply, which carries its calls, to the
conversation as received, decode each call with `call.decode(serializer)`,
and add one tool message (`toolCallId = call.id`) per call before sending
again. A model may call several tools at once; each call needs its result.

See [ToolCallingExample.kt](src/test/kotlin/ai/router/sdk/examples/ToolCallingExample.kt).

### Structured Output

Declare the response shape as a `@Serializable` class and set
`ResponseFormat(serializer)` as the settings' `responseFormat`. The SDK
generates the schema from the class and sends it; the response is an ordinary
response, which may also carry tool calls. `responseFormat.decode(response)`
reads the answer into the class, strictly. The schema carries exactly what the class declares. Annotate a
class with [`@NoAdditionalProperties`](src/main/kotlin/ai/router/sdk/schema/NoAdditionalProperties.kt)
to rule out properties beyond those declared; some providers require it.

See [StructuredOutputExample.kt](src/test/kotlin/ai/router/sdk/examples/StructuredOutputExample.kt).

### Reasoning

Every chat request names a reasoning effort (`none`, `low`, `medium` or
`high`); `none` turns reasoning off. With a reasoning-capable model and a level above `none`, read
the visible answer and the model's reasoning trace separately
(via `response.message.reasoningContent`).

See [ReasoningExample.kt](src/test/kotlin/ai/router/sdk/examples/ReasoningExample.kt).

### Embeddings

Each text of `input` corresponds to one embedding in the response. The
optional `dimensions` requests a specific embedding size from the provider.

See [EmbeddingsExample.kt](src/test/kotlin/ai/router/sdk/examples/EmbeddingsExample.kt).

### Web Search

Send a query and get web pages back, most relevant first. Each result has a
URL, a title and a text snippet. `maxResults` caps how many results come back.

See [SearchExample.kt](src/test/kotlin/ai/router/sdk/examples/SearchExample.kt).

### Page Contents

Load the content of one or more web pages. The response has one result per
requested URL, in request order. A page that fails to load carries an `error`
instead.

See [ContentsExample.kt](src/test/kotlin/ai/router/sdk/examples/ContentsExample.kt).

### Text to Speech

Turn a text into speech in one of a model's voices. Each text-to-speech model
entry lists the voices it speaks in; pass a voice's id with the text. The
response carries the audio as a base64-encoded WAV file (16-bit mono PCM).

See [TextToSpeechExample.kt](src/test/kotlin/ai/router/sdk/examples/TextToSpeechExample.kt).

### Speech to Text

Turn speech into text. The request carries the audio as a base64-encoded WAV
file of 16-bit mono PCM, at any sample rate: the format text-to-speech
returns, so a text-to-speech response's `audio` can be passed on as is.
Audio in any other format is rejected with a 400. The response carries the
transcript as `text`.

See [SpeechToTextExample.kt](src/test/kotlin/ai/router/sdk/examples/SpeechToTextExample.kt).

### Listing Models

Each use case lists the models its `send` accepts (e.g.
`client.chat.listModels()`). Entries carry metadata specific to their use
case; chat models, for example, report their features (tools, vision,
reasoning, structured output). Optional filters narrow a listing by
provider type (`cloud`/`local`) or a case-insensitive substring of the model
id. Each entry's `model` property is the fully-qualified string to pass
verbatim as a request's model.

See [ListModelsExample.kt](src/test/kotlin/ai/router/sdk/examples/ListModelsExample.kt).

### Error Handling

Non-2xx responses surface as `AiRouterException`, which carries the HTTP
status code and a message of the form `"<error_type>: <error_message>"`.

See [ErrorHandlingExample.kt](src/test/kotlin/ai/router/sdk/examples/ErrorHandlingExample.kt).

### Configuration

`AiRouterClient(baseUrl)` takes the server's base URL and nothing else: the
SDK brings its own HTTP setup, with a 10-second connect timeout and a
10-minute request timeout for slow model calls. To bound a call more tightly,
wrap it in `withTimeout`, which cancels the request.

## Testing

The example files in
[src/test/kotlin/ai/router/sdk/examples/](src/test/kotlin/ai/router/sdk/examples/)
are runnable JUnit tests that exercise every documented usage pattern against
a live ai-router server.

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
