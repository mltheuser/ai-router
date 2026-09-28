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
 * The routing identity every listed model carries, whatever its use case.
 *
 * [model] is the fully-qualified string (`id:provider_type@provider`) to
 * pass verbatim as a request's `model` to address this entry.
 */
public interface ModelRef {
    public val id: String
    public val model: String
    public val provider: String
    public val providerType: ProviderType
}

/**
 * The response of a use case's model listing (e.g. `GET /v1/chat/models`).
 */
@Serializable
public data class ModelList<M : ModelRef>(
    val data: List<M>,
)
