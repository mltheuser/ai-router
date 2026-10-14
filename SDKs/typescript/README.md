# ai-router TypeScript SDK

TypeScript SDK for the [ai-router](../../) proxy. Requests and responses are
plain objects typed after the wire format, sent with `fetch` to a local or
remote proxy. The SDK has no runtime dependencies and uses only web-platform
globals (`fetch`, `URL`, `atob`, …), so it runs wherever those exist: Node,
Bun or a browser.

Every documented usage pattern is backed by a runnable example in
[tests/](tests/).

The examples double as the SDK's integration test suite — see
[Testing](#testing) below.

## Installation

The SDK is not published. Depend on it by path and import it as TypeScript
source; your bundler (e.g. esbuild) compiles it along with your own code.

```json
"dependencies": {
  "ai-router-sdk": "file:../ai-router/SDKs/typescript"
}
```

Your tsconfig needs `moduleResolution` `bundler` (implied by `module`
`Preserve`), and the fetch types from either the `DOM` lib or `@types/node`.

## Usage

The client mirrors the server: one property per use case (e.g. `client.chat`,
`client.embedding`) — each with `send(...)` for the use case's request and
`listModels(...)` for the models it accepts.

Request and response types transcribe the server's wire format, snake_case
field names included, so a request is an object literal of the documented
JSON. Responses are typed but not validated: the SDK trusts the server.

### Quick Start

A minimal end-to-end usage of the SDK: connect to the proxy, send one chat
request, read the text response with `chatText`.

See [quickStart.test.ts](tests/quickStart.test.ts).

> Append `@<provider>` to a model string to pin a request to a specific
> backend (e.g. `llama3.2:cloud@openrouter`).

### Multi-Turn Conversation

Send a sequence of system / user / assistant messages and collect a final
response. Demonstrates request-level sampling parameters
(`temperature`, `max_tokens`, …).

See [multiTurn.test.ts](tests/multiTurn.test.ts).

### Vision (Multimodal)

Attach an image alongside a text prompt. The image is passed inline as
base64-encoded bytes with an explicit MIME type.

See [vision.test.ts](tests/vision.test.ts).

### Tool Calling

Define a tool whose parameters are a plain JSON Schema object. The example
walks through the full round-trip: issue the call, read the model's tool call
and its arguments, and send the tool result back as a follow-up turn.

See [toolCalling.test.ts](tests/toolCalling.test.ts).

### Structured Output

Describe the response shape as a JSON Schema in the request's
`response_format`, and parse the response with `parseStructuredOutput`. The
parsed JSON is cast to the type you name, not validated against the schema.

See [structuredOutput.test.ts](tests/structuredOutput.test.ts).

### Reasoning

Every chat request names a reasoning effort (`none`, `low`, `medium` or
`high`); `none` turns reasoning off. With a reasoning-capable model and a
level above `none`, read the visible answer and the model's reasoning trace
separately (via `chatText` and `chatReasoning`).

See [reasoning.test.ts](tests/reasoning.test.ts).

### Embeddings

Each text of `input` corresponds to one embedding in the response. The
optional `dimensions` requests a specific embedding size from the provider.

See [embeddings.test.ts](tests/embeddings.test.ts).

### Web Search

Send a query and get web pages back, most relevant first. Each result has a
URL, a title and a text snippet. `max_results` caps how many results come
back.

See [search.test.ts](tests/search.test.ts).

### Page Contents

Load the content of one or more web pages. The response has one result per
requested URL, in request order. A page that fails to load carries an `error`
instead.

See [contents.test.ts](tests/contents.test.ts).

### Text to Speech

Turn a text into speech in one of a model's voices. Each text-to-speech model
entry lists the voices it speaks in; pass a voice's id with the text. The
response carries the audio as a base64-encoded WAV file (16-bit mono PCM),
which `wavBytes` decodes.

See [textToSpeech.test.ts](tests/textToSpeech.test.ts).

### Listing Models

Each use case lists the models its `send` accepts (e.g.
`client.chat.listModels()`). Entries carry metadata specific to their use
case; chat models, for example, report their features (tools, vision,
reasoning, structured output). Optional filters narrow a listing by
provider type (`cloud`/`local`) or a case-insensitive substring of the model
id. Each entry's `model` property is the fully-qualified string to pass
verbatim as a request's model.

See [listModels.test.ts](tests/listModels.test.ts).

### Error Handling

Non-2xx responses reject with `AiRouterError`, which carries the HTTP
`status`, the decoded error body as `apiError` (`type`, `message` and an
optional `code`), and a message of the form `"<type>: <message>"`. A failure
of `fetch` itself, such as an unreachable server or an aborted request,
rejects with that failure as is.

See [errorHandling.test.ts](tests/errorHandling.test.ts).

### Configuration

Pass your own `fetch` to `new AiRouterClient({ baseUrl, fetch })` to take full
control of the HTTP layer: timeouts, logging, retries, etc. Without it, the
client uses the global `fetch`. The example tests in this repo all build
their client through a shared factory that demonstrates this pattern with a
timeout — see [setup.ts](tests/setup.ts).

## Testing

The example files in [tests/](tests/) are runnable `bun test` tests that
exercise every documented usage pattern against a live ai-router server.

Prerequisites:

- A running ai-router server. See the project's [testing guide](../../TESTING.md)
  for how to build and run one.
- A provider with the models referenced in [setup.ts](tests/setup.ts).
  Adjust the constants as needed.

Run all examples:

```bash
bun install
bun run test
```

## Linting

Static analysis runs through [Biome](https://biomejs.dev/), with every rule
enabled, over both the sources and the examples. The TypeScript compiler runs
with strict settings (`strict`, `exactOptionalPropertyTypes`,
`noUncheckedIndexedAccess`, …); the sources are checked against the web
platform alone, the examples against Bun's types.

```bash
bun run lint        # Biome: lint and format check
bun run fmt         # Biome: apply the safe fixes and formatting
bun run typecheck   # tsc
```

Config lives in [biome.jsonc](biome.jsonc), which documents every relaxed
rule inline, and in [tsconfig.json](tsconfig.json).
