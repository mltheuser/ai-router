package ai.router.sdk.models

import kotlinx.serialization.Serializable

/**
 * OpenAI-compatible API error body.
 */
@Serializable
public data class ApiError(
    val type: String,
    val message: String,
    val code: String? = null,
)

/**
 * Top-level error response wrapper.
 */
@Serializable
public data class ErrorResponse(
    val error: ApiError,
)

/**
 * Exception thrown when the ai-router returns a non-2xx response.
 */
public class AiRouterException(
    public val statusCode: Int,
    public val apiError: ApiError,
) : RuntimeException("${apiError.type}: ${apiError.message}")
