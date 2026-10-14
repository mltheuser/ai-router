import { expect, test } from "bun:test";
import { newExampleClient, SEARCH_MODEL } from "./setup";

test("Search: find web pages for a query, most relevant first", async () => {
  const client = newExampleClient();

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
