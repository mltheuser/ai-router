package ai.router.sdk.schema

import kotlinx.serialization.SerialInfo

/**
 * Associates a description with a serializable property.
 * This description is included in the generated JSON schema when
 * requesting structured output from an LLM.
 */
@OptIn(kotlinx.serialization.ExperimentalSerializationApi::class)
@SerialInfo
@Target(AnnotationTarget.PROPERTY, AnnotationTarget.CLASS)
annotation class Description(val value: String)
