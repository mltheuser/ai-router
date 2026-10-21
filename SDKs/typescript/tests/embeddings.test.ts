import { expect, test } from "bun:test";
import { AiRouterClient } from "ai-router-sdk";
import { EMBED_MODEL, SERVER_URL } from "./setup";

test("Embeddings: embed several texts in a single call", async () => {
  const client = new AiRouterClient(SERVER_URL);

  const response = await client.embedding.send({
    model: EMBED_MODEL,
    input: ["First document to embed", "Second document to embed", "Third document to embed"],
    dimensions: 768,
  });

  expect(response.data).toHaveLength(3);
  for (const item of response.data) {
    expect(item.embedding, `embedding ${item.index}`).not.toBeEmpty();
  }
});
