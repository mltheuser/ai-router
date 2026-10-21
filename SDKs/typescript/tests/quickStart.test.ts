import { expect, test } from "bun:test";
import { AiRouterClient, messageText } from "ai-router-sdk";
import { CHAT_EFFORT, CHAT_MODEL, SERVER_URL } from "./setup";

test("Quick start: send a chat request and read the response", async () => {
  const client = new AiRouterClient(SERVER_URL);

  const response = await client.chat.send({ model: CHAT_MODEL, reasoning_effort: CHAT_EFFORT }, [
    { role: "system", content: [{ type: "text", text: "You are a helpful assistant." }] },
    { role: "user", content: [{ type: "text", text: "What is the capital of France?" }] },
  ]);

  expect(messageText(response.message), `finish_reason=${response.finish_reason}`).not.toBeEmpty();
});
