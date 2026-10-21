import { expect, test } from "bun:test";
import { AiRouterClient, type ChatMessage, type ChatSettings, messageText } from "ai-router-sdk";
import { CHAT_EFFORT, CHAT_MODEL, SERVER_URL } from "./setup";

test("Multi-turn conversation with sampling parameters", async () => {
  const client = new AiRouterClient(SERVER_URL);
  const settings: ChatSettings = {
    model: CHAT_MODEL,
    reasoning_effort: CHAT_EFFORT,
    temperature: 0.7,
    max_tokens: 2048,
  };
  const conversation: ChatMessage[] = [
    { role: "system", content: [{ type: "text", text: "You are a helpful assistant." }] },
    { role: "user", content: [{ type: "text", text: "My name is Alice." }] },
  ];

  // Each reply joins the conversation as received, then the next question.
  conversation.push((await client.chat.send(settings, conversation)).message);
  conversation.push({ role: "user", content: [{ type: "text", text: "What's my name?" }] });
  const response = await client.chat.send(settings, conversation);

  console.log(`Answer: ${messageText(response.message)}`);
  expect(messageText(response.message), `finish_reason=${response.finish_reason}`).not.toBeEmpty();
});
