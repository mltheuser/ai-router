package ai.router.sdk.speechtotext

import ai.router.sdk.ModelList
import ai.router.sdk.ProviderType
import ai.router.sdk.Transport

/**
 * The speech-to-text use case: speech in, as a WAV file, its transcript out.
 */
public class SpeechToTextClient internal constructor(private val transport: Transport) {

    /**
     * Send a speech-to-text request. Audio that is not a WAV file of 16-bit
     * mono PCM is rejected with a 400 [ai.router.sdk.AiRouterException].
     */
    public suspend fun send(request: SpeechToTextRequest): SpeechToTextResponse = transport.send(USE_CASE, request)

    /**
     * List the models [send] accepts, optionally narrowed by provider [type]
     * and a case-insensitive [search] substring of the model id.
     */
    public suspend fun listModels(type: ProviderType? = null, search: String? = null): ModelList<SpeechToTextModel> =
        transport.listModels(USE_CASE, type, search)

    private companion object {
        const val USE_CASE = "speech-to-text"
    }
}
