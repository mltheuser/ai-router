package ai.router.sdk.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/**
 * OpenAI-compatible embedding request.
 */
@Serializable
public data class EmbedRequest(
    val model: String,
    val input: List<String>,
    val dimensions: Int? = null,
)

/**
 * OpenAI-compatible embedding response.
 */
@Serializable
public data class EmbedResponse(
    val `object`: String,
    val data: List<EmbedData>,
    val model: String,
    val usage: EmbedUsage,
)

@Serializable
public data class EmbedData(
    val `object`: String,
    val embedding: List<Double>,
    val index: Int,
)

@Serializable
public data class EmbedUsage(
    @SerialName("prompt_tokens") val promptTokens: Int,
    @SerialName("total_tokens") val totalTokens: Int,
)
