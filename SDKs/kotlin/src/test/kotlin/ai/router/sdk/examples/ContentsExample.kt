package ai.router.sdk.examples

import ai.router.sdk.contents.ContentsRequest
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test
import kotlin.test.assertEquals

class ContentsExample {

    @Test
    @DisplayName("Contents: load the content of a web page")
    fun run() = runBlocking {
        newExampleClient().use { client ->
            val response = client.contents.send(
                ContentsRequest(
                    model = CONTENTS_MODEL,
                    urls = listOf("https://www.rfc-editor.org/rfc/rfc2119"),
                )
            )

            // One result per requested URL: either the page or why it failed.
            assertEquals(1, response.results.size, "expected one result per requested URL")
            val page = response.results.single()
            when (val error = page.error) {
                null -> println("${page.title}: ${page.text?.length} characters")
                else -> println("Failed to load ${page.url}: $error")
            }
        }
    }
}
