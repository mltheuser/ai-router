package ai.router.sdk.examples

import ai.router.sdk.AiRouterClient
import io.ktor.client.HttpClient
import io.ktor.client.engine.cio.CIO
import io.ktor.client.plugins.HttpTimeout
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.serialization.kotlinx.json.json

// ─── Server + model targets ───────────────────────────────────────────

internal const val SERVER_URL = "http://localhost:8787"
internal const val CHAT_MODEL = "gemma4:31b-it-qat:local@ollama"
internal const val EMBED_MODEL = "qwen3-embedding:4b:local@ollama"

// ─── Shared client factory ────────────────────────────────────────────

internal fun newExampleClient(): AiRouterClient {
    val httpClient = HttpClient(CIO) {
        install(ContentNegotiation) { json() }
        install(HttpTimeout) {
            requestTimeoutMillis = 300_000
            connectTimeoutMillis = 10_000
        }
    }
    return AiRouterClient(SERVER_URL, httpClient)
}
