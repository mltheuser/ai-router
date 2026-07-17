package ai.router.sdk.models

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/**
 * Distinguishes cloud from local providers.
 */
@Serializable
public enum class ProviderType {
    @SerialName("cloud")
    CLOUD,

    @SerialName("local")
    LOCAL,
}

/**
 * Model capability flags.
 */
@Serializable
public enum class Capability {
    @SerialName("chat")
    CHAT,

    @SerialName("embed")
    EMBED,

    @SerialName("structured_output")
    STRUCTURED_OUTPUT,

    @SerialName("reasoning")
    REASONING,

    @SerialName("tools")
    TOOLS,

    @SerialName("vision")
    VISION,
}

/**
 * Describes a model available through a specific provider.
 *
 * [model] is the fully-qualified string (`id:provider_type@provider`) to
 * pass verbatim as a request's `model` to address this entry.
 */
@Serializable
public data class ModelInfo(
    val id: String,
    val model: String,
    val provider: String,
    @SerialName("provider_type") val providerType: ProviderType,
    val capabilities: List<Capability>,
    @SerialName("context_window") val contextWindow: Int = 0,
    @SerialName("cost_per_m_input") val costPerMInput: Double? = null,
    @SerialName("cost_per_m_output") val costPerMOutput: Double? = null,
    @SerialName("size_bytes") val sizeBytes: Long? = null,
) {
    public fun hasCapability(cap: Capability): Boolean = cap in capabilities
}

/**
 * Response format for listing models.
 */
@Serializable
public data class ModelList(
    val `object`: String,
    val data: List<ModelInfo>,
)
