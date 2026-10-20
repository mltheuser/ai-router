package ai.router.sdk.chat

import ai.router.sdk.ModelList
import ai.router.sdk.ProviderType
import ai.router.sdk.Transport

/**
 * The chat use case: a conversation in, the model's next message out.
 */
public class ChatClient internal constructor(private val transport: Transport) {

    /** The model's next message in the conversation [messages], answered as [settings] say. */
    public suspend fun send(settings: ChatSettings, messages: List<ChatMessage>): ChatResponse =
        transport.send(USE_CASE, ChatRequest(settings, messages))

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
