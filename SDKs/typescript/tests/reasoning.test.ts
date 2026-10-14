import { expect, test } from "bun:test";
import { chatReasoning, chatText } from "ai-router-sdk";
import { CHAT_MODEL, newExampleClient } from "./setup";

test("Reasoning: separate the reasoning trace from the final answer", async () => {
  const client = newExampleClient();

  const response = await client.chat.send({
    model: CHAT_MODEL,
    reasoning_effort: "high",
    messages: [{ role: "user", content: [{ type: "text", text: "Prove that √2 is irrational." }] }],
  });

  expect(chatText(response), `finish_reason=${response.finish_reason}`).not.toBeEmpty();
  expect(chatReasoning(response), "expected a reasoning trace").not.toBeEmpty();
});
