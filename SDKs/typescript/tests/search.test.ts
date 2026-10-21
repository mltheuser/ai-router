import { expect, test } from "bun:test";
import { AiRouterClient } from "ai-router-sdk";
import { SEARCH_MODEL, SERVER_URL } from "./setup";

test("Search: find web pages for a query, most relevant first", async () => {
  const client = new AiRouterClient(SERVER_URL);

  const response = await client.search.send({
    model: SEARCH_MODEL,
    query: "What is new in the latest Go release?",
    max_results: 5,
  });

  for (const result of response.results) {
    console.log(`${result.title} — ${result.url}`);
  }
  expect(response.results.length).toBeLessThanOrEqual(5);
});
