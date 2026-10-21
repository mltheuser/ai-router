import { expect, test } from "bun:test";
import { AiRouterClient } from "ai-router-sdk";
import { CONTENTS_MODEL, SERVER_URL } from "./setup";

test("Contents: load the content of a web page", async () => {
  const client = new AiRouterClient(SERVER_URL);

  const response = await client.contents.send({
    model: CONTENTS_MODEL,
    urls: ["https://www.rfc-editor.org/rfc/rfc2119"],
  });

  // One result per requested URL: either the page or why it failed.
  expect(response.results).toHaveLength(1);
  for (const page of response.results) {
    if (page.error === undefined) {
      console.log(`${page.title}: ${page.text?.length} characters`);
    } else {
      console.log(`Failed to load ${page.url}: ${page.error}`);
    }
  }
});
