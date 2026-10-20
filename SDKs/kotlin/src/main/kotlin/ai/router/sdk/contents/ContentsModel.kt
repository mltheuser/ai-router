package ai.router.sdk.contents

import ai.router.sdk.ModelRef
import ai.router.sdk.ProviderType
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/**
 * A contents model at one provider, as listed by `GET /v1/contents/models`:
 * one way the provider can load pages, e.g. from its cache or always fresh.
 */
@Serializable
public data class ContentsModel(
    override val id: String,
    override val model: String,
    override val provider: String,
    @SerialName("provider_type") override val providerType: ProviderType,
) : ModelRef
