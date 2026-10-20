package ai.router.sdk.embedding

import ai.router.sdk.ModelRef
import ai.router.sdk.ProviderType
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/**
 * An embedding model at one provider, as listed by `GET /v1/embedding/models`.
 *
 * The cost field is per million input tokens; `null` means unknown and `0.0` means free.
 */
@Serializable
public data class EmbeddingModel(
    override val id: String,
    override val model: String,
    override val provider: String,
    @SerialName("provider_type") override val providerType: ProviderType,
    @SerialName("context_window") val contextWindow: Int = 0,
    @SerialName("cost_per_m_input") val costPerMInput: Double? = null,
    @SerialName("size_bytes") val sizeBytes: Long? = null,
) : ModelRef
