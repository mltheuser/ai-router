package ai.router.sdk

import ai.router.sdk.chat.ChatClient
import ai.router.sdk.contents.ContentsClient
import ai.router.sdk.embedding.EmbeddingClient
import ai.router.sdk.search.SearchClient
import ai.router.sdk.speechtotext.SpeechToTextClient
import ai.router.sdk.texttospeech.TextToSpeechClient

/**
 * Client for the ai-router proxy, with one property per use case.
 *
 * ```kotlin
 * AiRouterClient("http://localhost:8787").use { client ->
 *     val models = client.chat.listModels()
 * }
 * ```
 *
 * The client brings its own HTTP setup: a 10-second connect timeout and a
 * 10-minute request timeout, long enough for slow model calls. Bound a call
 * more tightly with `withTimeout`, which cancels the request.
 *
 * @param baseUrl Root URL of the ai-router server (e.g. `http://localhost:8787`).
 */
public class AiRouterClient(baseUrl: String) : AutoCloseable {

    private val transport = Transport(baseUrl)

    public val chat: ChatClient = ChatClient(transport)
    public val embedding: EmbeddingClient = EmbeddingClient(transport)
    public val search: SearchClient = SearchClient(transport)
    public val contents: ContentsClient = ContentsClient(transport)
    public val textToSpeech: TextToSpeechClient = TextToSpeechClient(transport)
    public val speechToText: SpeechToTextClient = SpeechToTextClient(transport)

    override fun close() {
        transport.close()
    }
}
