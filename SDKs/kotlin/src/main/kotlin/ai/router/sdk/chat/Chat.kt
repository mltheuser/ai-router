package ai.router.sdk.chat

import ai.router.sdk.ModelRef
import ai.router.sdk.ProviderType
import ai.router.sdk.schema.SchemaGenerator
import kotlinx.serialization.KSerializer
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonObject

// ─── Models ───────────────────────────────────────────────────────────

/**
 * Optional abilities of a chat model beyond plain text chat.
 */
@Serializable
public enum class ChatFeature {
    @SerialName("tools")
    TOOLS,

    @SerialName("vision")
    VISION,

    @SerialName("reasoning")
    REASONING,

    @SerialName("structured_output")
    STRUCTURED_OUTPUT,
}

/**
 * A chat model at one provider, as listed by `GET /v1/chat/models`.
 *
 * Cost fields are per million tokens; `null` means unknown and `0.0` means free.
 */
@Serializable
public data class ChatModel(
    override val id: String,
    override val model: String,
    override val provider: String,
    @SerialName("provider_type") override val providerType: ProviderType,
    val features: List<ChatFeature> = emptyList(),
    @SerialName("context_window") val contextWindow: Int = 0,
    @SerialName("cost_per_m_input") val costPerMInput: Double? = null,
    @SerialName("cost_per_m_output") val costPerMOutput: Double? = null,
    @SerialName("size_bytes") val sizeBytes: Long? = null,
) : ModelRef {
    public fun has(vararg required: ChatFeature): Boolean = features.containsAll(required.toList())
}

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

// ─── Roles and finish reasons ─────────────────────────────────────────

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

// ─── Content parts (multimodal) ───────────────────────────────────────

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

// ─── Tool calling ─────────────────────────────────────────────────────

/**
 * A tool the model may call, its arguments described by the JSON Schema
 * [parameters], e.g. `SchemaGenerator.generate(Args.serializer())`.
 */
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

/** A tool invocation requested by the model; [id] matches a tool result to its call. */
@Serializable
public data class ToolCall(
    val id: String,
    val function: ToolCallFunction,
) {
    /**
     * The call's arguments, decoded with the serializer of the tool's
     * parameters. Decoding is strict and throws a `SerializationException`
     * on arguments that do not match them, e.g. a key they do not declare.
     */
    public fun <T> decode(serializer: KSerializer<T>): T =
        Json.decodeFromJsonElement(serializer, function.arguments)
}

// ─── Structured output ────────────────────────────────────────────────

/**
 * Asks the model to answer with JSON of [serializer]'s schema, sent as the
 * request's `response_format`. The schema carries exactly what the classes
 * declare; see [ai.router.sdk.schema.NoAdditionalProperties]. A response to
 * such a request is an ordinary response, which [decode] reads.
 */
public data class ResponseFormat<T>(
    val serializer: KSerializer<T>,
    val name: String = serializer.descriptor.serialName.substringAfterLast('.'),
    val description: String? = null,
) {
    /**
     * The text of [response]'s message, decoded into [T]. Decoding is strict
     * and throws a `SerializationException` if the text does not match the
     * schema, e.g. when the model called tools instead of answering, or was
     * cut off.
     */
    public fun decode(response: ChatResponse): T = Json.decodeFromString(serializer, response.message.text)

    internal fun toWire(): ResponseFormatWire = ResponseFormatWire(
        type = "json_schema",
        jsonSchema = JsonSchemaSpec(name, description, SchemaGenerator.generate(serializer)),
    )
}

@Serializable
internal data class ResponseFormatWire(
    val type: String,
    @SerialName("json_schema") val jsonSchema: JsonSchemaSpec,
)

@Serializable
internal data class JsonSchemaSpec(
    val name: String,
    val description: String? = null,
    val schema: JsonObject,
)

// ─── Messages ─────────────────────────────────────────────────────────

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

// ─── Request / Response ───────────────────────────────────────────────

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

/** The answer to a chat request: the model's next message. */
@Serializable
public data class ChatResponse(
    val model: String,
    val message: ChatMessage,
    @SerialName("finish_reason") val finishReason: FinishReason,
    val usage: ChatUsage,
)
