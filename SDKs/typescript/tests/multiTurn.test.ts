import { expect, test } from "bun:test";
import { chatText } from "ai-router-sdk";
import { CHAT_EFFORT, CHAT_MODEL, newExampleClient } from "./setup";

test("Multi-turn conversation with sampling parameters", async () => {
  const client = newExampleClient();

  const response = await client.chat.send({
    model: CHAT_MODEL,
    reasoning_effort: CHAT_EFFORT,
    messages: [
      { role: "system", content: [{ type: "text", text: "You are a helpful assistant." }] },
      { role: "user", content: [{ type: "text", text: "My name is Alice." }] },
      { role: "assistant", content: [{ type: "text", text: "Hello Alice! How can I help you?" }] },
      { role: "user", content: [{ type: "text", text: "What's my name?" }] },
    ],
    temperature: 0.7,
    max_tokens: 2048,
  });

  expect(chatText(response), `finish_reason=${response.finish_reason}`).not.toBeEmpty();
});
