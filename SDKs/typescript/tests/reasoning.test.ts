import { expect, test } from "bun:test";
import { AiRouterClient, messageText } from "ai-router-sdk";
import { CHAT_MODEL, SERVER_URL } from "./setup";

// A novel puzzle: models with adaptive thinking answer famous problems without
// thinking, so a canonical one shows no trace.
const PUZZLE =
  "Three friends - Ana, Ben and Cy - each have a different pet (cat, dog, fish) and " +
  "live in houses 1, 2 and 3. Ana is not in house 1. The dog owner is in house 2. Ben " +
  "owns the fish. Cy is not in house 3. Who owns the cat, and in which house?";

test("Reasoning: separate the reasoning trace from the final answer", async () => {
  const client = new AiRouterClient(SERVER_URL);

  const response = await client.chat.send({ model: CHAT_MODEL, reasoning_effort: "high" }, [
    { role: "user", content: [{ type: "text", text: PUZZLE }] },
  ]);

  expect(messageText(response.message), `finish_reason=${response.finish_reason}`).not.toBeEmpty();
  expect(response.message.reasoning_content, "expected a reasoning trace").not.toBeEmpty();
});
