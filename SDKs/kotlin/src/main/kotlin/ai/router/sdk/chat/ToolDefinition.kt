package ai.router.sdk.chat

import kotlinx.serialization.Serializable
import kotlinx.serialization.json.JsonObject

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
