package ai.router.sdk.texttospeech

import ai.router.sdk.ModelList
import ai.router.sdk.ProviderType
import ai.router.sdk.Transport

/**
 * The text-to-speech use case: a text and a voice in, the text spoken in that
 * voice out, as a WAV file.
 */
public class TextToSpeechClient internal constructor(private val transport: Transport) {

    public suspend fun send(request: TextToSpeechRequest): TextToSpeechResponse = transport.send(USE_CASE, request)

    /**
     * List the models [send] accepts, each with its voices, optionally
     * narrowed by provider [type] and a case-insensitive [search] substring
     * of the model id.
     */
    public suspend fun listModels(type: ProviderType? = null, search: String? = null): ModelList<TextToSpeechModel> =
        transport.listModels(USE_CASE, type, search)

    private companion object {
        const val USE_CASE = "text-to-speech"
    }
}
