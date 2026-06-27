package ai.router.sdk.schema

import kotlinx.serialization.ExperimentalSerializationApi
import kotlinx.serialization.descriptors.PrimitiveKind
import kotlinx.serialization.descriptors.SerialDescriptor
import kotlinx.serialization.descriptors.SerialKind
import kotlinx.serialization.descriptors.StructureKind
import kotlinx.serialization.json.JsonArray
import kotlinx.serialization.json.JsonElement
import kotlinx.serialization.json.JsonObject
import kotlinx.serialization.json.JsonPrimitive
import kotlinx.serialization.json.add
import kotlinx.serialization.json.buildJsonObject
import kotlinx.serialization.json.put
import kotlinx.serialization.json.putJsonArray
import kotlinx.serialization.json.putJsonObject
import kotlinx.serialization.serializer

/**
 * Derives a JSON Schema from a kotlinx-serialization [SerialDescriptor].
 *
 * The generated schema can be sent as the `response_format.json_schema.schema`
 * field so the LLM returns structured output matching a `@Serializable` data class.
 *
 * Supported mappings:
 * - STRING / CHAR → `{"type": "string"}`
 * - BOOLEAN       → `{"type": "boolean"}`
 * - BYTE/SHORT/INT/LONG → `{"type": "integer"}`
 * - FLOAT/DOUBLE  → `{"type": "number"}`
 * - ENUM          → `{"type": "string", "enum": [...]}`
 * - LIST          → `{"type": "array", "items": ...}`
 * - MAP           → `{"type": "object", "additionalProperties": ...}`
 * - CLASS/OBJECT  → `{"type": "object", "properties": {...}, "required": [...]}`
 */
@OptIn(ExperimentalSerializationApi::class)
public object SchemaGenerator {

    /**
     * Generate a JSON Schema [JsonObject] for the reified type [T].
     */
    public inline fun <reified T> generate(): JsonObject =
        generate(serializer<T>().descriptor)

    @PublishedApi
    internal fun generate(descriptor: SerialDescriptor): JsonObject =
        descriptorToSchema(descriptor)

    private fun descriptorToSchema(descriptor: SerialDescriptor): JsonObject {
        // Unwrap nullable wrapper — its single element is the real descriptor.
        val d = if (descriptor.isNullable && descriptor.elementsCount > 0) {
            descriptor.getElementDescriptor(0)
        } else {
            descriptor
        }

        return when (d.kind) {
            is PrimitiveKind -> primitiveSchema(d.kind as PrimitiveKind)

            // ── Enum ──────────────────────────────────────────────────
            SerialKind.ENUM -> buildJsonObject {
                put("type", "string")
                putJsonArray("enum") {
                    for (i in 0 until d.elementsCount) {
                        add(JsonPrimitive(d.getElementName(i)))
                    }
                }
            }

            // ── List / Array ──────────────────────────────────────────
            StructureKind.LIST -> buildJsonObject {
                put("type", "array")
                put("items", descriptorToSchema(d.getElementDescriptor(0)))
            }

            // ── Map ───────────────────────────────────────────────────
            StructureKind.MAP -> buildJsonObject {
                put("type", "object")
                // element 0 = key descriptor, element 1 = value descriptor
                put("additionalProperties", descriptorToSchema(d.getElementDescriptor(1)))
            }

            // ── Object / data class ───────────────────────────────────
            StructureKind.CLASS,
            StructureKind.OBJECT -> objectSchema(d)

            else -> buildJsonObject { put("type", "string") }
        }
    }

    private fun primitiveSchema(kind: PrimitiveKind): JsonObject = buildJsonObject {
        val type = when (kind) {
            PrimitiveKind.BOOLEAN -> "boolean"
            PrimitiveKind.BYTE,
            PrimitiveKind.SHORT,
            PrimitiveKind.INT,
            PrimitiveKind.LONG -> "integer"
            PrimitiveKind.FLOAT,
            PrimitiveKind.DOUBLE -> "number"
            else -> "string" // STRING, CHAR
        }
        put("type", type)
    }

    private fun objectSchema(d: SerialDescriptor): JsonObject = buildJsonObject {
        put("type", "object")
        val requiredNames = mutableListOf<JsonElement>()
        putJsonObject("properties") {
            for (i in 0 until d.elementsCount) {
                val name = d.getElementName(i)
                val elemDesc = d.getElementDescriptor(i)
                val descriptionAnn = d.getElementAnnotations(i)
                    .filterIsInstance<Description>().firstOrNull()

                val propSchema = descriptorToSchema(elemDesc)
                if (descriptionAnn != null) {
                    val map = propSchema.toMutableMap()
                    map["description"] = JsonPrimitive(descriptionAnn.value)
                    put(name, JsonObject(map))
                } else {
                    put(name, propSchema)
                }

                if (!d.isElementOptional(i) && !elemDesc.isNullable) {
                    requiredNames.add(JsonPrimitive(name))
                }
            }
        }
        if (requiredNames.isNotEmpty()) {
            put("required", JsonArray(requiredNames))
        }
    }
}
