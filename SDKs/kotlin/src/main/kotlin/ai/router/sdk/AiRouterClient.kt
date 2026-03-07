package ai.router.sdk

import ai.router.sdk.models.*
import io.ktor.client.*
import io.ktor.client.call.*
import io.ktor.client.engine.cio.*
import io.ktor.client.plugins.*
import io.ktor.client.plugins.contentnegotiation.*
import io.ktor.client.request.*
import io.ktor.client.statement.*
import io.ktor.http.*
import io.ktor.serialization.kotlinx.json.*
import kotlinx.serialization.KSerializer
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
class AiRouterClient(
    private val baseUrl: String,
    private val httpClient: HttpClient = defaultHttpClient(),
) : AutoCloseable {

    companion object {
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
                requestTimeoutMillis = 600_000
                connectTimeoutMillis = 10_000
            }
        }
    }

    // ─── Public API ───────────────────────────────────────────────────

    /**
     * Send a chat completion request.
     */
    suspend fun chat(request: ChatRequest): ChatResponse {
        return post("/v1/chat/completions", request)
    }

    /**
     * Send a chat completion request with structured output and deserialize
     * the response content directly into [T].
     *
     * The [ChatRequest] should have been built with `structuredOutput<T>()` in the DSL
     * so the server returns JSON matching the schema of [T].
     *
     * @param deserializer The [KSerializer] for [T] (e.g. `MyClass.serializer()`).
     */
    suspend fun <T> chat(request: ChatRequest, deserializer: KSerializer<T>): T {
        val response = chat(request)
        val text = response.textContent
        return json.decodeFromString(deserializer, text)
    }

    /**
     * Send an embedding request.
     */
    suspend fun embed(request: EmbedRequest): EmbedResponse {
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
