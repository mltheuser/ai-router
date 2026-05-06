package ai.router.sdk.schema

import kotlinx.serialization.ExperimentalSerializationApi
import kotlinx.serialization.descriptors.*
import kotlinx.serialization.json.*
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
object SchemaGenerator {

    /**
     * Generate a JSON Schema [JsonObject] for the reified type [T].
     */
    inline fun <reified T> generate(): JsonObject =
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
            // ── Primitives ────────────────────────────────────────────
            PrimitiveKind.STRING,
            PrimitiveKind.CHAR -> buildJsonObject {
                put("type", "string")
            }

            PrimitiveKind.BOOLEAN -> buildJsonObject {
                put("type", "boolean")
            }

            PrimitiveKind.BYTE,
            PrimitiveKind.SHORT,
            PrimitiveKind.INT,
            PrimitiveKind.LONG -> buildJsonObject {
                put("type", "integer")
            }

            PrimitiveKind.FLOAT,
            PrimitiveKind.DOUBLE -> buildJsonObject {
                put("type", "number")
            }

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
            StructureKind.OBJECT -> buildJsonObject {
                put("type", "object")
                val requiredNames = mutableListOf<JsonElement>()
                putJsonObject("properties") {
                    for (i in 0 until d.elementsCount) {
                        val name = d.getElementName(i)
                        val elemDesc = d.getElementDescriptor(i)
                        val annotations = d.getElementAnnotations(i)
                        val descriptionAnn = annotations.filterIsInstance<Description>().firstOrNull()

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

            else -> buildJsonObject { put("type", "string") }
        }
    }
}
