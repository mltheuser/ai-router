package ai.router.sdk.chat

import ai.router.sdk.schema.SchemaGenerator
import kotlinx.serialization.KSerializer
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import kotlinx.serialization.json.Json
import kotlinx.serialization.json.JsonObject

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
