package ai.router.sdk.embedding

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/** The answer to an embedding request: one [Embedding] per input text. */
@Serializable
public data class EmbeddingResponse(
    val model: String,
    val data: List<Embedding>,
    val usage: EmbeddingUsage,
)

/** The vector of the input text at [index]. */
@Serializable
public data class Embedding(
    val index: Int,
    val embedding: List<Double>,
)

@Serializable
public data class EmbeddingUsage(
    @SerialName("prompt_tokens") val promptTokens: Int,
    @SerialName("total_tokens") val totalTokens: Int,
)
