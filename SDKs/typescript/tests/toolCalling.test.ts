import { expect, test } from "bun:test";
import { AiRouterClient, type ChatMessage, chatText, type ToolDefinition } from "ai-router-sdk";
import { CHAT_EFFORT, CHAT_MODEL, SERVER_URL } from "./setup";

// A tool's parameters are a plain JSON Schema object.
const getWeather: ToolDefinition = {
  name: "get_weather",
  description: "Get current weather for a city",
  parameters: {
    type: "object",
    properties: {
      city: { type: "string", description: "The city to check weather for" },
      unit: { type: "string", enum: ["celsius", "fahrenheit"] },
    },
    required: ["city"],
  },
};

test("Tool calling: define a tool, read the call, send the result back", async () => {
  const client = new AiRouterClient(SERVER_URL);
  const question: ChatMessage = {
    role: "user",
    content: [{ type: "text", text: "What's the weather in Berlin?" }],
  };

  // The model should answer with a call of the tool.
  const first = await client.chat.send({
    model: CHAT_MODEL,
    reasoning_effort: CHAT_EFFORT,
    messages: [question],
    tools: [getWeather],
  });

  const toolCall = first.message.tool_calls?.[0];
  if (toolCall === undefined) {
    throw new Error(`model produced no tool call (finish_reason=${first.finish_reason})`);
  }
  expect(toolCall.function.arguments.city).toBeString();

  // Send the assistant's message back as is, followed by the tool's result.
  const followUp = await client.chat.send({
    model: CHAT_MODEL,
    reasoning_effort: CHAT_EFFORT,
    messages: [
      question,
      first.message,
      {
        role: "tool",
        tool_call_id: toolCall.id,
        content: [{ type: "text", text: '{"temp_celsius": 22, "condition": "sunny"}' }],
      },
    ],
    tools: [getWeather],
  });

  expect(chatText(followUp), `finish_reason=${followUp.finish_reason}`).not.toBeEmpty();
});
