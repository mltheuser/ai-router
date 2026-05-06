package ai.router.sdk.schema

import kotlinx.serialization.SerialInfo

/**
 * Associates a description with a serializable property or class.
 * The description is included in the generated JSON schema — used for both
 * structured output response types and tool parameter classes.
 */
@OptIn(kotlinx.serialization.ExperimentalSerializationApi::class)
@SerialInfo
@Target(AnnotationTarget.PROPERTY, AnnotationTarget.CLASS)
annotation class Description(val value: String)
