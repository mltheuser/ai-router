package ai.router.sdk.search

import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

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
