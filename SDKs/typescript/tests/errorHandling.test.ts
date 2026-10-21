import { expect, test } from "bun:test";
import { AiRouterClient, AiRouterError } from "ai-router-sdk";
import { SERVER_URL } from "./setup";

test("Error handling: a non-2xx response rejects with AiRouterError, a cancelled call with its abort reason", async () => {
  const client = new AiRouterClient(SERVER_URL);

  const error = await client.chat
    .send({ model: "does-not-exist:local@nonexistent", reasoning_effort: "none" }, [
      { role: "user", content: [{ type: "text", text: "hi" }] },
    ])
    .catch((reason: unknown) => reason);

  if (!(error instanceof AiRouterError)) {
    throw new Error(`expected an AiRouterError, got ${String(error)}`);
  }
  expect(error.status).toBeWithin(400, 600);
  expect(error.apiError.type).not.toBeEmpty();
  expect(error.apiError.message).not.toBeEmpty();

  // Every call takes an optional signal; this one is cancelled before it starts.
  const cancelled = await client.chat
    .listModels({}, { signal: AbortSignal.abort() })
    .catch((reason: unknown) => reason);

  expect(cancelled).toBeInstanceOf(DOMException);
  expect((cancelled as DOMException).name).toBe("AbortError");
});
