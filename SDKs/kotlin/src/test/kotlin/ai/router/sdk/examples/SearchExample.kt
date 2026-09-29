package ai.router.sdk.examples

import ai.router.sdk.search.SearchRequest
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test
import kotlin.test.assertTrue

class SearchExample {

    @Test
    @DisplayName("Search: find web pages for a query, most relevant first")
    fun run() = runBlocking {
        newExampleClient().use { client ->
            val response = client.search.send(
                SearchRequest(
                    model = SEARCH_MODEL,
                    query = "What is new in the latest Go release?",
                    maxResults = 5,
                )
            )

            response.results.forEach { println("${it.title} — ${it.url}") }
            assertTrue(response.results.size <= 5, "expected at most the requested 5 results")
        }
    }
}
