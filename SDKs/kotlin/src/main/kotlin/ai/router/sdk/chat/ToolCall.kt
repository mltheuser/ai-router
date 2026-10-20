package ai.router.sdk.chat

import kotlinx.serialization.KSerializer
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonObject

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

@Serializable
public data class ToolCallFunction(
    val name: String,
    val arguments: JsonObject,
)
