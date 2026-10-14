import { expect, test } from "bun:test";
import { CONTENTS_MODEL, newExampleClient } from "./setup";

test("Contents: load the content of a web page", async () => {
  const client = newExampleClient();

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
