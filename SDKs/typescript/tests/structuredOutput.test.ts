import { expect, test } from "bun:test";
import { AiRouterClient, ResponseFormat } from "ai-router-sdk";
import { z } from "zod";
import { CHAT_EFFORT, CHAT_MODEL, SERVER_URL } from "./setup";

// What some providers require: every object rules out further properties, and
// every property is required, a missing value being null.
const WeatherInfo = z.strictObject({
  city: z.string(),
  temp_celsius: z.number(),
  condition: z.string().nullable(),
});

test("Structured output: decode the answer into a typed, validated object", async () => {
  const client = new AiRouterClient(SERVER_URL);
  const weatherFormat = new ResponseFormat(WeatherInfo, { name: "weather_info" });

  const response = await client.chat.send(
    { model: CHAT_MODEL, reasoning_effort: CHAT_EFFORT, response_format: weatherFormat },
    [
      {
        role: "system",
        content: [{ type: "text", text: "Extract weather information from the text." }],
      },
      { role: "user", content: [{ type: "text", text: "It's 22°C and sunny in Berlin today." }] },
    ],
  );

  const weather = weatherFormat.decode(response);

  expect(weather.city, JSON.stringify(weather)).not.toBeEmpty();
});
