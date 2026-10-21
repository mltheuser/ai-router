import { expect, test } from "bun:test";
import {
  AiRouterClient,
  type ChatMessage,
  type ChatSettings,
  decodeToolCall,
  messageText,
  toJsonSchema,
} from "ai-router-sdk";
import { z } from "zod";
import { CHAT_EFFORT, CHAT_MODEL, SERVER_URL } from "./setup";

// One schema describes the tool's parameters to the model and decodes its calls.
const WeatherParams = z.object({
  city: z.string().describe("The city to check weather for"),
  unit: z.enum(["celsius", "fahrenheit"]).default("celsius"),
});

test("Tool calling: define a tool, decode the calls, send the results back", async () => {
  const client = new AiRouterClient(SERVER_URL);
  const settings: ChatSettings = {
    model: CHAT_MODEL,
    reasoning_effort: CHAT_EFFORT,
    tools: [
      {
        name: "get_weather",
        description: "Get current weather for a city",
        parameters: toJsonSchema(WeatherParams),
      },
    ],
  };
  const conversation: ChatMessage[] = [
    { role: "user", content: [{ type: "text", text: "What's the weather in Berlin?" }] },
  ];

  // The reply carries the model's calls; it joins the conversation as received.
  const reply = (await client.chat.send(settings, conversation)).message;
  conversation.push(reply);
  const calls = reply.tool_calls ?? [];
  expect(calls, "model produced no tool call").not.toBeEmpty();

  // Every call needs its result before the conversation can continue.
  for (const call of calls) {
    const params = decodeToolCall(call, WeatherParams);
    expect(params.city).not.toBeEmpty();
    conversation.push({
      role: "tool",
      tool_call_id: call.id,
      content: [{ type: "text", text: '{"temp_celsius": 22, "condition": "sunny"}' }],
    });
  }
  const followUp = await client.chat.send(settings, conversation);

  console.log(`Answer: ${messageText(followUp.message)}`);
  expect(messageText(followUp.message), `finish_reason=${followUp.finish_reason}`).not.toBeEmpty();
});
