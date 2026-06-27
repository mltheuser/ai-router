package ai.router.sdk.dsl

import ai.router.sdk.models.ChatMessage
import ai.router.sdk.models.ChatRequest
import ai.router.sdk.models.ContentPart
import ai.router.sdk.models.ContentPartType
import ai.router.sdk.models.JsonSchemaSpec
import ai.router.sdk.models.ReasoningEffort
import ai.router.sdk.models.ResponseFormat
import ai.router.sdk.models.ResponseFormatType
import ai.router.sdk.models.StructuredChatRequest
import ai.router.sdk.models.ToolDefinition
import ai.router.sdk.schema.SchemaGenerator
import kotlinx.serialization.serializer

/**
 * Entry point for building a [ChatRequest] via DSL.
 *
 * ```kotlin
 * val request = chatRequest("llama3.2:local") {
 *     messages {
 *         system { text("You are a helpful assistant.") }
 *         user { text("Hello!") }
 *     }
 *     temperature(0.7)
 * }
 * ```
 */
public fun chatRequest(model: String, block: ChatRequestBuilder.() -> Unit): ChatRequest {
    return ChatRequestBuilder(model).apply(block).build()
}

/**
 * Entry point for a chat request with typed structured output.
 *
 * Sets up the response format from [T]'s schema automatically and pairs the
 * request with its deserializer so [ai.router.sdk.AiRouterClient.chat] returns
 * a [T] directly.
 *
 * ```kotlin
 * val request = structuredChatRequest<WeatherInfo>("gpt-4:cloud") {
 *     messages {
 *         system { text("Extract weather info.") }
 *         user { text("It's 22°C and sunny in Berlin.") }
 *     }
 * }
 * val weather: WeatherInfo = client.chat(request)
 * ```
 */
public inline fun <reified T> structuredChatRequest(
    model: String,
    block: ChatRequestBuilder.() -> Unit = {},
): StructuredChatRequest<T> {
    val builder = ChatRequestBuilder(model).apply(block)
    builder.applyResponseFormat(
        ResponseFormat(
            type = ResponseFormatType.JSON_SCHEMA,
            jsonSchema = JsonSchemaSpec(
                name = T::class.simpleName ?: "response",
                schema = SchemaGenerator.generate<T>(),
            ),
        )
    )
    return StructuredChatRequest(builder.build(), serializer<T>())
}

// ─── ChatRequest builder ──────────────────────────────────────────────

@DslMarker
public annotation class ChatDsl

@ChatDsl
public class ChatRequestBuilder(private val model: String) {
    private var messagesBuilder: MessagesBuilder? = null
    private var temperature: Double? = null
    private var maxTokens: Int? = null
    private var topP: Double? = null
    private var frequencyPenalty: Double? = null
    private var presencePenalty: Double? = null
    private var reasoningEffort: ReasoningEffort? = null
    private var responseFormat: ResponseFormat? = null
    private val tools = mutableListOf<ToolDefinition>()

    public fun messages(block: MessagesBuilder.() -> Unit) {
        messagesBuilder = MessagesBuilder().apply(block)
    }

    public fun temperature(value: Double) { temperature = value }
    public fun maxTokens(value: Int) { maxTokens = value }
    public fun topP(value: Double) { topP = value }
    public fun frequencyPenalty(value: Double) { frequencyPenalty = value }
    public fun presencePenalty(value: Double) { presencePenalty = value }
    public fun reasoningEffort(value: ReasoningEffort) { reasoningEffort = value }

    public fun tools(block: ToolsBuilder.() -> Unit) {
        tools.addAll(ToolsBuilder().apply(block).build())
    }

    @PublishedApi
    internal fun applyResponseFormat(format: ResponseFormat) {
        responseFormat = format
    }

    @PublishedApi
    internal fun build(): ChatRequest {
        val msgs = messagesBuilder?.build()
            ?: error("messages { } block is required")
        return ChatRequest(
            model = model,
            messages = msgs,
            temperature = temperature,
            maxTokens = maxTokens,
            topP = topP,
            frequencyPenalty = frequencyPenalty,
            presencePenalty = presencePenalty,
            reasoningEffort = reasoningEffort?.name?.lowercase(),
            responseFormat = responseFormat,
            tools = tools.ifEmpty { null },
        )
    }
}

// ─── Messages builder ─────────────────────────────────────────────────

@ChatDsl
public class MessagesBuilder {
    private val messages = mutableListOf<ChatMessage>()

    public fun system(block: ContentBuilder.() -> Unit) {
        messages.add(buildMessage("system", block))
    }

    public fun user(block: ContentBuilder.() -> Unit) {
        messages.add(buildMessage("user", block))
    }

    public fun assistant(block: ContentBuilder.() -> Unit) {
        messages.add(buildMessage("assistant", block))
    }

    /**
     * Add a tool result message.
     *
     * ```kotlin
     * tool(callId = "call_abc") {
     *     text("{\"result\": 42}")
     * }
     * ```
     */
    public fun tool(callId: String, block: ContentBuilder.() -> Unit) {
        val parts = ContentBuilder().apply(block).build()
        messages.add(ChatMessage(role = "tool", content = parts, toolCallId = callId))
    }

    private fun buildMessage(role: String, block: ContentBuilder.() -> Unit): ChatMessage {
        val parts = ContentBuilder().apply(block).build()
        return ChatMessage(role = role, content = parts)
    }

    internal fun build(): List<ChatMessage> = messages.toList()
}

// ─── Content parts builder ────────────────────────────────────────────

@ChatDsl
public class ContentBuilder {
    private val parts = mutableListOf<ContentPart>()

    public fun text(value: String) {
        parts.add(ContentPart(type = ContentPartType.TEXT, text = value))
    }

    public fun image(mimeType: String, base64Data: String) {
        parts.add(ContentPart(type = ContentPartType.IMAGE, mimeType = mimeType, base64Data = base64Data))
    }

    internal fun build(): List<ContentPart> = parts.toList()
}

// ─── Tools builder ────────────────────────────────────────────────────

@ChatDsl
public class ToolsBuilder {
    private val tools = mutableListOf<ToolDefinition>()

    /**
     * Define a tool using a `@Serializable` class for its parameter schema.
     *
     * ```kotlin
     * @Serializable
     * data class WeatherParams(
     *     @Description("The city to check") val city: String,
     *     val unit: String = "celsius",
     * )
     *
     * tools {
     *     tool<WeatherParams>("get_weather", "Get current weather for a city")
     * }
     * ```
     *
     * Decode the arguments from a tool call response with [ai.router.sdk.models.ToolCall.decode].
     */
    public inline fun <reified T> tool(name: String, description: String? = null) {
        addTool(ToolDefinition(name = name, description = description, parameters = SchemaGenerator.generate<T>()))
    }

    @PublishedApi
    internal fun addTool(tool: ToolDefinition) { tools.add(tool) }

    internal fun build(): List<ToolDefinition> = tools.toList()
}
