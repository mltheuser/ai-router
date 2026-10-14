import { expect, test } from "bun:test";
import { AiRouterError } from "ai-router-sdk";
import { newExampleClient } from "./setup";

test("Error handling: non-2xx responses reject with AiRouterError", async () => {
  const client = newExampleClient();

  const error = await client.chat
    .send({
      model: "does-not-exist:local@nonexistent",
      reasoning_effort: "none",
      messages: [{ role: "user", content: [{ type: "text", text: "hi" }] }],
    })
    .catch((reason: unknown) => reason);

  if (!(error instanceof AiRouterError)) {
    throw new Error(`expected an AiRouterError, got ${String(error)}`);
  }
  expect(error.status).toBeWithin(400, 600);
  expect(error.apiError.type).not.toBeEmpty();
  expect(error.apiError.message).not.toBeEmpty();
});
