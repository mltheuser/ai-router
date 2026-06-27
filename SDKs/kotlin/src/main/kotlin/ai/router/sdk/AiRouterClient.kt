package ai.router.sdk

import ai.router.sdk.models.AiRouterException
import ai.router.sdk.models.ApiError
import ai.router.sdk.models.ChatRequest
import ai.router.sdk.models.ChatResponse
import ai.router.sdk.models.EmbedRequest
import ai.router.sdk.models.EmbedResponse
import ai.router.sdk.models.ErrorResponse
import ai.router.sdk.models.StructuredChatRequest
import io.ktor.client.HttpClient
import io.ktor.client.call.body
import io.ktor.client.engine.cio.CIO
import io.ktor.client.plugins.HttpTimeout
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.client.request.post
import io.ktor.client.request.setBody
import io.ktor.client.statement.HttpResponse
import io.ktor.client.statement.bodyAsText
import io.ktor.http.ContentType
import io.ktor.http.contentType
import io.ktor.http.isSuccess
import io.ktor.serialization.kotlinx.json.json
import kotlinx.serialization.json.Json

/**
 * Client for the ai-router LLM proxy.
 *
 * ```kotlin
 * val client = AiRouterClient("http://localhost:8787")
 * client.use {
 *     val response = it.chat(request)
 *     println(response.textContent)
 * }
 * ```
 *
 * @param baseUrl Root URL of the ai-router server (e.g. `http://localhost:8787`).
 * @param httpClient Optional pre-configured Ktor [HttpClient]. If not provided, a default CIO client is created.
 */
public class AiRouterClient(
    private val baseUrl: String,
    private val httpClient: HttpClient = defaultHttpClient(),
) : AutoCloseable {

    private companion object {
        // LLM calls can run long; allow up to 10 minutes per request.
        private const val REQUEST_TIMEOUT_MILLIS = 600_000L
        private const val CONNECT_TIMEOUT_MILLIS = 10_000L

        private val json = Json {
            ignoreUnknownKeys = true
            encodeDefaults = false
            isLenient = true
        }

        private fun defaultHttpClient(): HttpClient = HttpClient(CIO) {
            install(ContentNegotiation) {
                json(json)
            }
            install(HttpTimeout) {
                requestTimeoutMillis = REQUEST_TIMEOUT_MILLIS
                connectTimeoutMillis = CONNECT_TIMEOUT_MILLIS
            }
        }
    }

    // ─── Public API ───────────────────────────────────────────────────

    /**
     * Send a chat completion request.
     */
    public suspend fun chat(request: ChatRequest): ChatResponse {
        return post("/v1/chat/completions", request)
    }

    /**
     * Send a structured chat completion request and deserialize the response
     * directly into [T]. Build the request with `structuredChatRequest<T>()`.
     */
    public suspend fun <T> chat(request: StructuredChatRequest<T>): T {
        val response = chat(request.inner)
        return json.decodeFromString(request.serializer, response.textContent)
    }

    /**
     * Send an embedding request.
     */
    public suspend fun embed(request: EmbedRequest): EmbedResponse {
        return post("/v1/embeddings", request)
    }

    override fun close() {
        httpClient.close()
    }

    // ─── Internal ─────────────────────────────────────────────────────

    private suspend inline fun <reified Req, reified Res> post(path: String, body: Req): Res {
        val response: HttpResponse = httpClient.post("$baseUrl$path") {
            contentType(ContentType.Application.Json)
            setBody(body)
        }
        if (!response.status.isSuccess()) {
            val errorBody = try {
                response.body<ErrorResponse>()
            } catch (_: Exception) {
                ErrorResponse(ApiError(type = "unknown", message = response.bodyAsText()))
            }
            throw AiRouterException(response.status.value, errorBody.error)
        }
        return response.body()
    }
}
