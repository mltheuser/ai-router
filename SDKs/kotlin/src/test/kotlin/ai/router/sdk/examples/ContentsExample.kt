package ai.router.sdk.examples

import ai.router.sdk.models.ContentsRequest
import ai.router.sdk.models.SearchRequest
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test
import kotlin.test.assertEquals

class ContentsExample {

    @Test
    @DisplayName("Contents: search, then load the top result's page")
    fun run() = runBlocking {
        newExampleClient().use { client ->
            // The agent flow: search first, then read the most promising page.
            val found = client.search(SearchRequest(model = SEARCH_MODEL, query = "Kotlin coroutines guide"))
            val topUrl = found.results.firstOrNull()?.url ?: "https://kotlinlang.org/docs/coroutines-guide.html"

            val response = client.getContents(ContentsRequest(model = CONTENTS_MODEL, urls = listOf(topUrl)))

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
