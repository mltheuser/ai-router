import { expect, test } from "bun:test";
import { chatReasoning, chatText } from "ai-router-sdk";
import { CHAT_MODEL, newExampleClient } from "./setup";

test("Reasoning: separate the reasoning trace from the final answer", async () => {
  const client = newExampleClient();

  const response = await client.chat.send({
    model: CHAT_MODEL,
    reasoning_effort: "high",
    // A novel puzzle: models with adaptive thinking answer famous problems
    // without thinking, so a canonical one shows no trace.
    messages: [
      {
        role: "user",
        content: [
          {
            type: "text",
            text:
              "Three friends - Ana, Ben and Cy - each have a different pet (cat, dog, fish) and " +
              "live in houses 1, 2 and 3. Ana is not in house 1. The dog owner is in house 2. Ben " +
              "owns the fish. Cy is not in house 3. Who owns the cat, and in which house?",
          },
        ],
      },
    ],
  });

  expect(chatText(response), `finish_reason=${response.finish_reason}`).not.toBeEmpty();
  expect(chatReasoning(response), "expected a reasoning trace").not.toBeEmpty();
});
