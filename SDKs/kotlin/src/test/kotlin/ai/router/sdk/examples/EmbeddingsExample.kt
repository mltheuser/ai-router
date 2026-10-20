package ai.router.sdk.examples

import ai.router.sdk.AiRouterClient
import ai.router.sdk.embedding.EmbeddingRequest
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class EmbeddingsExample {

    @Test
    @DisplayName("Embeddings: request multiple texts in a single call")
    fun run() = runBlocking {
        AiRouterClient(SERVER_URL).use { client ->
            val response = client.embedding.send(
                EmbeddingRequest(
                    model = EMBED_MODEL,
                    input = listOf("First document to embed", "Second document to embed", "Third document to embed"),
                    dimensions = 768,
                )
            )

            assertEquals(3, response.data.size, "expected one embedding per input text")
            response.data.forEach { item ->
                assertTrue(
                    item.embedding.isNotEmpty(),
                    "embedding ${item.index} was empty",
                )
            }
        }
    }
}
