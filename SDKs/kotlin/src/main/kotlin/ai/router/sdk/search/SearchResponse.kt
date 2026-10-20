package ai.router.sdk.search

import kotlinx.serialization.Serializable

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
