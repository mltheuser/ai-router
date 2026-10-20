package ai.router.sdk.chat

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/** Everything of a chat request but its messages: which model answers, and how. */
public data class ChatSettings(
    val model: String,
    val reasoningEffort: ReasoningEffort,
    val tools: List<ToolDefinition> = emptyList(),
    val responseFormat: ResponseFormat<*>? = null,
    val temperature: Double? = null,
    val topP: Double? = null,
    val maxTokens: Int? = null,
    val frequencyPenalty: Double? = null,
    val presencePenalty: Double? = null,
)

@Serializable
public enum class ReasoningEffort {
    @SerialName("none")
    NONE,

    @SerialName("low")
    LOW,

    @SerialName("medium")
    MEDIUM,

    @SerialName("high")
    HIGH,
}
