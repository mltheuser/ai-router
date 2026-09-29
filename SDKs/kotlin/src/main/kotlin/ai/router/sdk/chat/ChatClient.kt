package ai.router.sdk.chat

import ai.router.sdk.ModelList
import ai.router.sdk.ProviderType
import ai.router.sdk.Transport

/**
 * The chat use case: a conversation in, the model's next message out.
 */
public class ChatClient internal constructor(private val transport: Transport) {

    /** Send a chat request. Build one with [chatRequest]. */
    public suspend fun send(request: ChatRequest): ChatResponse = transport.send(USE_CASE, request)

    /**
     * Send a structured chat request and decode the response directly into
     * [T]. Build one with [structuredChatRequest].
     */
    public suspend fun <T> send(request: StructuredChatRequest<T>): T =
        Transport.json.decodeFromString(request.serializer, send(request.inner).textContent)

    /**
     * List the models [send] accepts, optionally narrowed by provider [type]
     * and a case-insensitive [search] substring of the model id.
     */
    public suspend fun listModels(type: ProviderType? = null, search: String? = null): ModelList<ChatModel> =
        transport.listModels(USE_CASE, type, search)

    private companion object {
        const val USE_CASE = "chat"
    }
}
