package ai.router.sdk.embedding

import ai.router.sdk.ModelList
import ai.router.sdk.ProviderType
import ai.router.sdk.Transport

/**
 * The embedding use case: texts in, one vector per text out.
 */
public class EmbeddingClient internal constructor(private val transport: Transport) {

    /** Send an embedding request. Build one with [embedRequest]. */
    public suspend fun send(request: EmbedRequest): EmbedResponse = transport.send(USE_CASE, request)

    /**
     * List the models [send] accepts, optionally narrowed by provider [type]
     * and a case-insensitive [search] substring of the model id.
     */
    public suspend fun listModels(type: ProviderType? = null, search: String? = null): ModelList<EmbedModel> =
        transport.listModels(USE_CASE, type, search)

    private companion object {
        const val USE_CASE = "embedding"
    }
}
