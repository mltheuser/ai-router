import { expect, test } from "bun:test";
import { AiRouterClient, parseStructuredOutput } from "ai-router-sdk";
import { CHAT_EFFORT, CHAT_MODEL, SERVER_URL } from "./setup";

interface WeatherInfo {
  city: string;
  temp_celsius: number;
  condition: string;
}

test("Structured output: parse the response into a typed object", async () => {
  const client = new AiRouterClient(SERVER_URL);

  const response = await client.chat.send({
    model: CHAT_MODEL,
    reasoning_effort: CHAT_EFFORT,
    messages: [
      {
        role: "system",
        content: [{ type: "text", text: "Extract weather information from the text." }],
      },
      { role: "user", content: [{ type: "text", text: "It's 22°C and sunny in Berlin today." }] },
    ],
    response_format: {
      type: "json_schema",
      json_schema: {
        name: "weather_info",
        schema: {
          type: "object",
          properties: {
            city: { type: "string" },
            temp_celsius: { type: "number" },
            condition: { type: "string" },
          },
          required: ["city", "temp_celsius", "condition"],
          additionalProperties: false,
        },
      },
    },
  });

  const weather = parseStructuredOutput<WeatherInfo>(response);

  expect(weather.city, JSON.stringify(weather)).not.toBeEmpty();
});
