package ai.router.sdk.models

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

/**
 * Page contents request: the pages to load.
 */
@Serializable
public data class ContentsRequest(
    val model: String,
    val urls: List<String>,
)

/**
 * Page contents response: one [ContentsResult] per requested URL, in request
 * order.
 */
@Serializable
public data class ContentsResponse(
    val model: String,
    val results: List<ContentsResult>,
)

/**
 * The content of one requested page. [url] is the URL as requested and
 * [text] the page as markdown. A page that could not be loaded has [error]
 * set, the provider's reason, and no [title] or [text]; the request as a
 * whole still succeeds.
 */
@Serializable
public data class ContentsResult(
    val url: String,
    val title: String? = null,
    val text: String? = null,
    val error: String? = null,
)
