import { expect, test } from "bun:test";
import { chatText } from "ai-router-sdk";
import { file } from "bun";
import { CHAT_EFFORT, CHAT_MODEL, newExampleClient } from "./setup";

test("Vision: send an image alongside a text prompt", async () => {
  const image = await file(new URL("resources/apple.png", import.meta.url)).bytes();
  const client = newExampleClient();

  const response = await client.chat.send({
    model: CHAT_MODEL,
    reasoning_effort: CHAT_EFFORT,
    messages: [
      {
        role: "user",
        content: [
          { type: "text", text: "Briefly describe this image." },
          { type: "image", mime_type: "image/png", base64_data: image.toBase64() },
        ],
      },
    ],
  });

  expect(chatText(response), `finish_reason=${response.finish_reason}`).not.toBeEmpty();
});
