package ai.router.sdk.contents

import ai.router.sdk.ModelList
import ai.router.sdk.ProviderType
import ai.router.sdk.Transport

/**
 * The contents use case: web page URLs in, each page's content out.
 * Reached as [ai.router.sdk.AiRouterClient.contents].
 */
public class ContentsClient internal constructor(private val transport: Transport) {

    /**
     * Send a contents request. The response has one result per requested
     * URL, in request order; a page that fails to load carries an error
     * instead of failing the whole request.
     */
    public suspend fun send(request: ContentsRequest): ContentsResponse = transport.send(USE_CASE, request)

    /**
     * List the models [send] accepts, optionally narrowed by provider [type]
     * and a case-insensitive [search] substring of the model id.
     */
    public suspend fun listModels(type: ProviderType? = null, search: String? = null): ModelList<ContentsModel> =
        transport.listModels(USE_CASE, type, search)

    private companion object {
        const val USE_CASE = "contents"
    }
}
