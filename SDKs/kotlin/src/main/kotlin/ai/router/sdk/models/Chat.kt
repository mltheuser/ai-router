package ai.router.sdk.models

import kotlinx.serialization.KSerializer
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.decodeFromJsonElement

// ─── Reasoning effort levels ──────────────────────────────────────────

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

// ─── Content parts (multimodal) ───────────────────────────────────────

@Serializable
public enum class ContentPartType {
    @SerialName("text")
    TEXT,

    @SerialName("image")
    IMAGE,
}

/**
 * One piece of a multimodal message.
 */
@Serializable
public data class ContentPart(
    val type: ContentPartType,
    val text: String? = null,
    @SerialName("mime_type") val mimeType: String? = null,
    @SerialName("base64_data") val base64Data: String? = null,
)

// ─── Tool calling ─────────────────────────────────────────────────────

@Serializable
public data class ToolDefinition(
    val name: String,
    val description: String? = null,
    val parameters: JsonObject? = null,
)

@Serializable
public data class ToolCallFunction(
    val name: String,
    val arguments: JsonObject,
)

@Serializable
public data class ToolCall(
    val id: String,
    val function: ToolCallFunction,
)

@PublishedApi
internal val toolCallJson: Json = Json { ignoreUnknownKeys = true }

public inline fun <reified T> ToolCall.decode(): T =
    toolCallJson.decodeFromJsonElement(function.arguments)

// ─── Response format / structured output ──────────────────────────────

@Serializable
public enum class ResponseFormatType {
    @SerialName("json_schema")
    JSON_SCHEMA,
}

@Serializable
public data class JsonSchemaSpec(
    val name: String,
    val description: String? = null,
    val schema: JsonObject? = null,
)

@Serializable
public data class ResponseFormat(
    val type: ResponseFormatType,
    @SerialName("json_schema") val jsonSchema: JsonSchemaSpec? = null,
)

// ─── Structured chat request ──────────────────────────────────────────

public data class StructuredChatRequest<T>(
    val inner: ChatRequest,
    val serializer: KSerializer<T>,
)

// ─── Messages ─────────────────────────────────────────────────────────

@Serializable
public data class ChatMessage(
    // Plain String, not a closed enum: role is a passthrough value (the Go server models it as
    // an open string type), so an enum would throw on any unlisted value during deserialization.
    val role: String,
    val content: List<ContentPart>,
    @SerialName("reasoning_content") val reasoningContent: String? = null,
    @SerialName("tool_calls") val toolCalls: List<ToolCall>? = null,
    @SerialName("tool_call_id") val toolCallId: String? = null,
)

// ─── Request / Response ───────────────────────────────────────────────

@Serializable
public data class ChatRequest(
    val model: String,
    val messages: List<ChatMessage>,
    @SerialName("frequency_penalty") val frequencyPenalty: Double? = null,
    @SerialName("max_tokens") val maxTokens: Int? = null,
    @SerialName("presence_penalty") val presencePenalty: Double? = null,
    val temperature: Double? = null,
    @SerialName("top_p") val topP: Double? = null,
    @SerialName("response_format") val responseFormat: ResponseFormat? = null,
    @SerialName("reasoning_effort") val reasoningEffort: ReasoningEffort? = null,
    val tools: List<ToolDefinition>? = null,
)

@Serializable
public data class ChatUsage(
    @SerialName("prompt_tokens") val promptTokens: Int,
    @SerialName("completion_tokens") val completionTokens: Int,
    @SerialName("total_tokens") val totalTokens: Int,
    @SerialName("reasoning_tokens") val reasoningTokens: Int = 0,
)

@Serializable
public data class ChatResponse(
    val model: String,
    val message: ChatMessage,
    // Plain String, not a closed enum: finish_reason is a passthrough value (providers may return
    // "error" or native values), so an enum would throw on an unlisted value during deserialization.
    @SerialName("finish_reason") val finishReason: String,
    val usage: ChatUsage,
) {
    /** Convenience: the text content of the response message. */
    public val textContent: String
        get() = message.content
            .filter { it.type == ContentPartType.TEXT }
            .mapNotNull { it.text }
            .joinToString("")
}
