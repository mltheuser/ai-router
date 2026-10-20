package ai.router.sdk.chat

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/** The body of `POST /v1/chat`. */
@Serializable
internal data class ChatRequest(
    val model: String,
    val messages: List<ChatMessage>,
    @SerialName("frequency_penalty") val frequencyPenalty: Double? = null,
    @SerialName("max_tokens") val maxTokens: Int? = null,
    @SerialName("presence_penalty") val presencePenalty: Double? = null,
    val temperature: Double? = null,
    @SerialName("top_p") val topP: Double? = null,
    @SerialName("response_format") val responseFormat: ResponseFormatWire? = null,
    @SerialName("reasoning_effort") val reasoningEffort: ReasoningEffort,
    val tools: List<ToolDefinition>? = null,
) {
    constructor(settings: ChatSettings, messages: List<ChatMessage>) : this(
        model = settings.model,
        messages = messages,
        frequencyPenalty = settings.frequencyPenalty,
        maxTokens = settings.maxTokens,
        presencePenalty = settings.presencePenalty,
        temperature = settings.temperature,
        topP = settings.topP,
        responseFormat = settings.responseFormat?.toWire(),
        reasoningEffort = settings.reasoningEffort,
        tools = settings.tools.ifEmpty { null },
    )
}
