package ai.router.sdk.speechtotext

import ai.router.sdk.ModelRef
import ai.router.sdk.ProviderType
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable
import java.util.Base64

/**
 * A speech-to-text model at one provider, as listed by
 * `GET /v1/speech-to-text/models`.
 *
 * [costPerAudioHour] is the price per hour of input audio; `null` means
 * unknown and `0.0` means free.
 */
@Serializable
public data class SpeechToTextModel(
    override val id: String,
    override val model: String,
    override val provider: String,
    @SerialName("provider_type") override val providerType: ProviderType,
    @SerialName("cost_per_audio_hour") val costPerAudioHour: Double? = null,
) : ModelRef

/**
 * Speech-to-text request. [audio] is a base64-encoded WAV file of 16-bit mono
 * PCM, at any sample rate: the format text-to-speech returns, so a
 * text-to-speech response's `audio` can be passed as is. [fromWav] encodes the
 * bytes of a WAV file.
 */
@Serializable
public data class SpeechToTextRequest(
    val model: String,
    val audio: String,
) {
    public companion object {
        /** A request to transcribe [wav], the bytes of a WAV file of 16-bit mono PCM. */
        public fun fromWav(model: String, wav: ByteArray): SpeechToTextRequest =
            SpeechToTextRequest(model, Base64.getEncoder().encodeToString(wav))
    }
}

/**
 * Speech-to-text response: [text] is the transcript of the whole audio.
 */
@Serializable
public data class SpeechToTextResponse(
    val model: String,
    val text: String,
)
