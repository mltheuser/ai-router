package ai.router.sdk.search

import ai.router.sdk.ModelRef
import ai.router.sdk.ProviderType
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/**
 * A search model at one provider, as listed by `GET /v1/search/models`: one
 * way the provider can search, e.g. a faster or a more thorough mode.
 */
@Serializable
public data class SearchModel(
    override val id: String,
    override val model: String,
    override val provider: String,
    @SerialName("provider_type") override val providerType: ProviderType,
) : ModelRef

/**
 * Web search request. [maxResults], if set, caps the number of results; the
 * provider may return fewer.
 */
@Serializable
public data class SearchRequest(
    val model: String,
    val query: String,
    @SerialName("max_results") val maxResults: Int? = null,
)

/**
 * Web search response. [results] are ordered by relevance, most relevant first.
 */
@Serializable
public data class SearchResponse(
    val model: String,
    val results: List<SearchResult>,
)

/**
 * One web page found for the query. [title] may be empty: not every page has
 * one. [snippet] holds the excerpts the provider selected as relevant to the
 * query; it is not necessarily short.
 */
@Serializable
public data class SearchResult(
    val url: String,
    val title: String,
    val snippet: String,
)
