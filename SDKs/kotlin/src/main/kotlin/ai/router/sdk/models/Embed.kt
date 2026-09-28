package ai.router.sdk.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/**
 * An embedding model at one provider, as listed by `GET /v1/embedding/models`.
 *
 * The cost field is per million input tokens; `null` means unknown and `0.0` means free.
 */
@Serializable
public data class EmbedModel(
    override val id: String,
    override val model: String,
    override val provider: String,
    @SerialName("provider_type") override val providerType: ProviderType,
    @SerialName("context_window") val contextWindow: Int = 0,
    @SerialName("cost_per_m_input") val costPerMInput: Double? = null,
    @SerialName("size_bytes") val sizeBytes: Long? = null,
) : ModelRef

/**
 * Embedding request.
 */
@Serializable
public data class EmbedRequest(
    val model: String,
    val input: List<String>,
    val dimensions: Int? = null,
)

/**
 * Embedding response: one [EmbedData] per input text.
 */
@Serializable
public data class EmbedResponse(
    val model: String,
    val data: List<EmbedData>,
    val usage: EmbedUsage,
)

/**
 * The vector of the input text at [index].
 */
@Serializable
public data class EmbedData(
    val index: Int,
    val embedding: List<Double>,
)

@Serializable
public data class EmbedUsage(
    @SerialName("prompt_tokens") val promptTokens: Int,
    @SerialName("total_tokens") val totalTokens: Int,
)
