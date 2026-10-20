package ai.router.sdk.chat

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/**
 * A message of a conversation, which is a `List<ChatMessage>`. A caller
 * writes system, user and tool messages; assistant messages are the model's
 * replies, added to the conversation as received.
 */
@Serializable
public data class ChatMessage(
    val role: ChatRole,
    val content: List<ContentPart>,
    /** The model's reasoning trace, on an assistant message. */
    @SerialName("reasoning_content") val reasoningContent: String? = null,
    /** The tools the model calls, on an assistant message; each needs a tool message answering it. */
    @SerialName("tool_calls") val toolCalls: List<ToolCall> = emptyList(),
    /** The [ToolCall.id] a tool message answers. */
    @SerialName("tool_call_id") val toolCallId: String? = null,
) {
    /** The message's text parts, concatenated. */
    public val text: String
        get() = content.filterIsInstance<ContentPart.Text>().joinToString("") { it.text }
}

/** The author of a message: the four roles the router accepts. */
@Serializable
public enum class ChatRole {
    @SerialName("system")
    SYSTEM,

    @SerialName("user")
    USER,

    @SerialName("assistant")
    ASSISTANT,

    @SerialName("tool")
    TOOL,
}

/** One piece of a message: a [Text] or an [Image]. */
@Serializable
public sealed interface ContentPart {

    // The router omits the text of an empty text part.
    @Serializable
    @SerialName("text")
    public data class Text(val text: String = "") : ContentPart

    /** An image inline as base64 without a data-URI prefix, e.g. of [mimeType] `image/png`. */
    @Serializable
    @SerialName("image")
    public data class Image(
        @SerialName("mime_type") val mimeType: String,
        @SerialName("base64_data") val base64Data: String,
    ) : ContentPart
}
