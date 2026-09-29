package ai.router.sdk.schema

import kotlinx.serialization.SerialInfo

/**
 * Associates a description with a serializable property or class; it is
 * emitted into the generated JSON schema.
 */
@OptIn(kotlinx.serialization.ExperimentalSerializationApi::class)
@SerialInfo
@Target(AnnotationTarget.PROPERTY, AnnotationTarget.CLASS)
public annotation class Description(val value: String)
