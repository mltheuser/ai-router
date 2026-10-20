package ai.router.sdk.contents

import kotlinx.serialization.Serializable

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
 * [text] the page's content, in a format up to the provider; a provider that
 * keeps the page's images includes them inline in [text], as links. A page
 * that could not be loaded has [error] set, the provider's reason, and no
 * [title] or [text]; the request as a whole still succeeds.
 */
@Serializable
public data class ContentsResult(
    val url: String,
    val title: String? = null,
    val text: String? = null,
    val error: String? = null,
)
