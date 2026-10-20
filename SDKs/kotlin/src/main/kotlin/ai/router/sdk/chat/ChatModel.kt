package ai.router.sdk.chat

import ai.router.sdk.ModelRef
import ai.router.sdk.ProviderType
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/**
 * A chat model at one provider, as listed by `GET /v1/chat/models`.
 *
 * Cost fields are per million tokens; `null` means unknown and `0.0` means free.
 */
@Serializable
public data class ChatModel(
    override val id: String,
    override val model: String,
    override val provider: String,
    @SerialName("provider_type") override val providerType: ProviderType,
    val features: List<ChatFeature> = emptyList(),
    @SerialName("context_window") val contextWindow: Int = 0,
    @SerialName("cost_per_m_input") val costPerMInput: Double? = null,
    @SerialName("cost_per_m_output") val costPerMOutput: Double? = null,
    @SerialName("size_bytes") val sizeBytes: Long? = null,
) : ModelRef {
    public fun has(vararg required: ChatFeature): Boolean = features.containsAll(required.toList())
}

/**
 * Optional abilities of a chat model beyond plain text chat.
 */
@Serializable
public enum class ChatFeature {
    @SerialName("tools")
    TOOLS,

    @SerialName("vision")
    VISION,

    @SerialName("reasoning")
    REASONING,

    @SerialName("structured_output")
    STRUCTURED_OUTPUT,
}
