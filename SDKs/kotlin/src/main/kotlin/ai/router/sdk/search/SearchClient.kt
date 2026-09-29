package ai.router.sdk.search

import ai.router.sdk.ModelList
import ai.router.sdk.ProviderType
import ai.router.sdk.Transport

/**
 * The search use case: a query in, a ranked list of web pages out.
 */
public class SearchClient internal constructor(private val transport: Transport) {

    /** Send a search request. */
    public suspend fun send(request: SearchRequest): SearchResponse = transport.send(USE_CASE, request)

    /**
     * List the models [send] accepts, optionally narrowed by provider [type]
     * and a case-insensitive [search] substring of the model id.
     */
    public suspend fun listModels(type: ProviderType? = null, search: String? = null): ModelList<SearchModel> =
        transport.listModels(USE_CASE, type, search)

    private companion object {
        const val USE_CASE = "search"
    }
}
