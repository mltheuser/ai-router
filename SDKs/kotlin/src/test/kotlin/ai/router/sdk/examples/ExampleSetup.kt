package ai.router.sdk.examples

import ai.router.sdk.AiRouterClient
import ai.router.sdk.chat.ReasoningEffort
import io.ktor.client.HttpClient
import io.ktor.client.engine.cio.CIO
import io.ktor.client.plugins.HttpTimeout
import io.ktor.client.plugins.contentnegotiation.ContentNegotiation
import io.ktor.serialization.kotlinx.json.json

// ─── Server + model targets ───────────────────────────────────────────

internal const val SERVER_URL = "http://localhost:8787"
internal const val CHAT_MODEL = "claude-sonnet-5-5:cloud@anthropic"

internal val CHAT_EFFORT = ReasoningEffort.LOW
internal const val EMBED_MODEL = "qwen3-embedding:4b:local@ollama"
internal const val SEARCH_MODEL = "fast:cloud@exa"
internal const val CONTENTS_MODEL = "auto:cloud@exa"
internal const val TEXT_TO_SPEECH_MODEL = "eleven_flash_v2_5:cloud@elevenlabs"
internal const val SPEECH_TO_TEXT_MODEL = "scribe_v2:cloud@elevenlabs"

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
