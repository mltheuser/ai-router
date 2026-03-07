package ai.router.sdk.models

import kotlinx.serialization.Serializable

/**
 * OpenAI-compatible API error body.
 */
@Serializable
data class ApiError(
    val type: String,
    val message: String,
    val code: String? = null,
)

/**
 * Top-level error response wrapper.
 */
@Serializable
data class ErrorResponse(
    val error: ApiError,
)

/**
 * Exception thrown when the ai-router returns a non-2xx response.
 */
class AiRouterException(
    val statusCode: Int,
    val apiError: ApiError,
) : RuntimeException("${apiError.type}: ${apiError.message}")
