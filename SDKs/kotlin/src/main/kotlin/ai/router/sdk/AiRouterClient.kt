package ai.router.sdk

import ai.router.sdk.chat.ChatClient
import ai.router.sdk.contents.ContentsClient
import ai.router.sdk.embedding.EmbeddingClient
import ai.router.sdk.search.SearchClient
import io.ktor.client.HttpClient
import io.ktor.client.engine.cio.CIO
import io.ktor.client.plugins.HttpTimeout
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.serialization.kotlinx.json.json

/**
 * Client for the ai-router proxy, with one property per use case.
 *
 * ```kotlin
 * AiRouterClient("http://localhost:8787").use { client ->
 *     val response = client.chat.send(request)
 *     println(response.textContent)
 * }
 * ```
 *
 * @param baseUrl Root URL of the ai-router server (e.g. `http://localhost:8787`).
 * @param httpClient Optional pre-configured Ktor [HttpClient] with JSON content
 * negotiation installed. If not provided, a default CIO client is created.
 */
public class AiRouterClient(
    baseUrl: String,
    httpClient: HttpClient = defaultHttpClient(),
) : AutoCloseable {

    private val transport = Transport(baseUrl, httpClient)

    public val chat: ChatClient = ChatClient(transport)
    public val embedding: EmbeddingClient = EmbeddingClient(transport)
    public val search: SearchClient = SearchClient(transport)
    public val contents: ContentsClient = ContentsClient(transport)

    override fun close() {
        transport.close()
    }

    private companion object {
        // LLM calls can run long; allow up to 10 minutes per request.
        private const val REQUEST_TIMEOUT_MILLIS = 600_000L
        private const val CONNECT_TIMEOUT_MILLIS = 10_000L

        private fun defaultHttpClient(): HttpClient = HttpClient(CIO) {
            install(ContentNegotiation) {
                json(Transport.json)
            }
            install(HttpTimeout) {
                requestTimeoutMillis = REQUEST_TIMEOUT_MILLIS
                connectTimeoutMillis = CONNECT_TIMEOUT_MILLIS
            }
        }
    }
}
