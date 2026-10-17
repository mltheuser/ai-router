import { expect, test } from "bun:test";
import type { ModelRef } from "ai-router-sdk";
import { newExampleClient } from "./setup";

test("List models: fetch each use case's models, with and without filters", async () => {
  const client = newExampleClient();

  // Every use case lists its own models, each with metadata of its own: chat
  // models report their features, for example.
  const chatModels = await client.chat.listModels();
  const embeddingModels = await client.embedding.listModels();
  const searchModels = await client.search.listModels();
  const contentsModels = await client.contents.listModels();
  const textToSpeechModels = await client.textToSpeech.listModels();
  const speechToTextModels = await client.speechToText.listModels();

  const allModels: ModelRef[] = [
    ...chatModels.data,
    ...embeddingModels.data,
    ...searchModels.data,
    ...contentsModels.data,
    ...textToSpeechModels.data,
    ...speechToTextModels.data,
  ];
  for (const entry of allModels) {
    expect(entry.model).toStartWith(`${entry.id}:`);
    expect(entry.model).toEndWith(`@${entry.provider}`);
  }

  // Pick a model by what it can do.
  const toolUser = chatModels.data.find(
    (entry) => entry.features?.includes("tools") && entry.features.includes("vision"),
  );
  console.log(`A chat model with tools and vision: ${toolUser?.model}`);

  // Narrow a listing by provider type and a case-insensitive id search.
  const [first] = chatModels.data;
  if (first === undefined) {
    throw new Error("expected a configured server to expose at least one chat model");
  }
  const fragment = first.id.slice(0, 4);
  const narrowed = await client.chat.listModels({ type: first.provider_type, search: fragment });

  expect(narrowed.data).not.toBeEmpty();
  for (const entry of narrowed.data) {
    expect(entry.id.toLowerCase()).toContain(fragment.toLowerCase());
  }

  const local = await client.embedding.listModels({ type: "local" });
  for (const entry of local.data) {
    expect(entry.provider_type).toBe("local");
  }
});
