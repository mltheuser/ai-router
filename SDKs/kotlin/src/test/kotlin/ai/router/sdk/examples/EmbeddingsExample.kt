package ai.router.sdk.examples

import ai.router.sdk.dsl.embedRequest
import kotlinx.coroutines.runBlocking
import kotlin.test.assertEquals
import kotlin.test.assertTrue
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test

class EmbeddingsExample {

    @Test
    @DisplayName("Embeddings: request multiple texts in a single call")
    fun run() = runBlocking {
        newExampleClient().use { client ->
            val response = client.embed(embedRequest(EMBED_MODEL) {
                batch("First document to embed")
                batch("Second document to embed")
                batch("Third document to embed")
                dimensions(768)
            })

            assertEquals(3, response.data.size, "expected one embedding per batch")
            response.data.forEach { item ->
                assertTrue(
                    item.embedding.isNotEmpty(),
                    "embedding ${item.index} was empty",
                )
            }
        }
    }
}
