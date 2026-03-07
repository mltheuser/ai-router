package ai.router.sdk.dsl

import ai.router.sdk.models.*
import ai.router.sdk.schema.SchemaGenerator
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonPrimitive
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
fun chatRequest(model: String, block: ChatRequestBuilder.() -> Unit): ChatRequest {
    return ChatRequestBuilder(model).apply(block).build()
}

// ─── ChatRequest builder ──────────────────────────────────────────────

@DslMarker
annotation class ChatDsl

@ChatDsl
class ChatRequestBuilder(private val model: String) {
    private var messagesBuilder: MessagesBuilder? = null
    private var temperature: Double? = null
    private var maxTokens: Int? = null
    private var topP: Double? = null
    private var frequencyPenalty: Double? = null
    private var presencePenalty: Double? = null
    private var reasoningEffort: ReasoningEffort? = null
    @PublishedApi internal var responseFormat: ResponseFormat? = null
    private val tools = mutableListOf<ToolDefinition>()

    fun messages(block: MessagesBuilder.() -> Unit) {
        messagesBuilder = MessagesBuilder().apply(block)
    }

    fun temperature(value: Double) { temperature = value }
    fun maxTokens(value: Int) { maxTokens = value }
    fun topP(value: Double) { topP = value }
    fun frequencyPenalty(value: Double) { frequencyPenalty = value }
    fun presencePenalty(value: Double) { presencePenalty = value }
    fun reasoningEffort(value: ReasoningEffort) { reasoningEffort = value }

    /**
     * Request structured output whose schema is derived automatically
     * from the `@Serializable` class [T].
     *
     * Pair this with `client.chat(request, T.serializer())` to get typed responses.
     */
    inline fun <reified T> structuredOutput(
        name: String = T::class.simpleName ?: "response",
        description: String? = null,
    ) {
        val descriptor = serializer<T>().descriptor
        val schema = SchemaGenerator.generate(descriptor)
        responseFormat = ResponseFormat(
            type = ResponseFormatType.JSON_SCHEMA,
            jsonSchema = JsonSchemaSpec(
                name = name,
                description = description,
                schema = schema,
            ),
        )
    }

    fun tools(block: ToolsBuilder.() -> Unit) {
        tools.addAll(ToolsBuilder().apply(block).build())
    }

    internal fun build(): ChatRequest {
        val msgs = messagesBuilder?.build()
            ?: throw IllegalStateException("messages { } block is required")
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
class MessagesBuilder {
    private val messages = mutableListOf<ChatMessage>()

    fun system(block: ContentBuilder.() -> Unit) {
        messages.add(buildMessage("system", block))
    }

    fun user(block: ContentBuilder.() -> Unit) {
        messages.add(buildMessage("user", block))
    }

    fun assistant(block: ContentBuilder.() -> Unit) {
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
    fun tool(callId: String, block: ContentBuilder.() -> Unit) {
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
class ContentBuilder {
    private val parts = mutableListOf<ContentPart>()

    fun text(value: String) {
        parts.add(ContentPart(type = ContentPartType.TEXT, text = value))
    }

    fun image(mimeType: String, base64Data: String) {
        parts.add(ContentPart(type = ContentPartType.IMAGE, mimeType = mimeType, base64Data = base64Data))
    }

    internal fun build(): List<ContentPart> = parts.toList()
}

// ─── Tools builder ────────────────────────────────────────────────────

@ChatDsl
class ToolsBuilder {
    private val tools = mutableListOf<ToolDefinition>()

    /**
     * Define a tool the model may call.
     *
     * @param name Tool function name.
     * @param description Human-readable description.
     * @param parametersBlock Lambda to build the JSON schema for parameters.
     */
    fun tool(
        name: String,
        description: String? = null,
        parametersBlock: (MutableMap<String, JsonElement>.() -> Unit)? = null,
    ) {
        val params = if (parametersBlock != null) {
            val map = mutableMapOf<String, JsonElement>()
            map.parametersBlock()
            map
        } else {
            null
        }
        tools.add(ToolDefinition(name = name, description = description, parameters = params))
    }

    internal fun build(): List<ToolDefinition> = tools.toList()
}
