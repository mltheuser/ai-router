package ai.router.sdk.search

import ai.router.sdk.ModelRef
import ai.router.sdk.ProviderType
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/**
 * A search model at one provider, as listed by `GET /v1/search/models`: one
 * way the provider can search, e.g. a fast mode or a deep research mode.
 */
@Serializable
public data class SearchModel(
    override val id: String,
    override val model: String,
    override val provider: String,
    @SerialName("provider_type") override val providerType: ProviderType,
) : ModelRef
