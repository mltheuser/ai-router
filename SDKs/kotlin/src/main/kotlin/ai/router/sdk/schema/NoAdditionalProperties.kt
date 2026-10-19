package ai.router.sdk.schema

import kotlinx.serialization.SerialInfo

/**
 * Emits `"additionalProperties": false` into the JSON schema of the annotated
 * class: an object of it has its declared properties and no others. Some
 * providers require this of every object in a structured-output schema; a
 * nested class needs its own annotation.
 */
@OptIn(kotlinx.serialization.ExperimentalSerializationApi::class)
@SerialInfo
@Target(AnnotationTarget.CLASS)
public annotation class NoAdditionalProperties
