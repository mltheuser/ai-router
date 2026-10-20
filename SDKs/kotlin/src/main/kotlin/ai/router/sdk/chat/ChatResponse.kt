package ai.router.sdk.chat

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/** The answer to a chat request: the model's next message. */
@Serializable
public data class ChatResponse(
    val model: String,
    val message: ChatMessage,
    @SerialName("finish_reason") val finishReason: FinishReason,
    val usage: ChatUsage,
)

/**
 * Why the model stopped. The constants name the values the router produces;
 * providers may return others, which decode as is.
 */
@Serializable
@JvmInline
public value class FinishReason(public val value: String) {
    override fun toString(): String = value

    public companion object {
        public val STOP: FinishReason = FinishReason("stop")
        public val TOOL_CALLS: FinishReason = FinishReason("tool_calls")
        public val LENGTH: FinishReason = FinishReason("length")
        public val CONTENT_FILTER: FinishReason = FinishReason("content_filter")
        public val ERROR: FinishReason = FinishReason("error")
    }
}

/** The tokens a request consumed. */
@Serializable
public data class ChatUsage(
    /** The whole prompt, cached tokens included. */
    @SerialName("prompt_tokens") val promptTokens: Int,
    @SerialName("completion_tokens") val completionTokens: Int,
    @SerialName("total_tokens") val totalTokens: Int,
    @SerialName("reasoning_tokens") val reasoningTokens: Int,
    @SerialName("cache_read_tokens") val cacheReadTokens: Int,
) {
    /** The tokens of both requests together, e.g. to total a conversation. */
    public operator fun plus(other: ChatUsage): ChatUsage = ChatUsage(
        promptTokens = promptTokens + other.promptTokens,
        completionTokens = completionTokens + other.completionTokens,
        totalTokens = totalTokens + other.totalTokens,
        reasoningTokens = reasoningTokens + other.reasoningTokens,
        cacheReadTokens = cacheReadTokens + other.cacheReadTokens,
    )

    public companion object {
        public val ZERO: ChatUsage = ChatUsage(0, 0, 0, 0, 0)
    }
}
