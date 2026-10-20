package ai.router.sdk.texttospeech

import ai.router.sdk.ModelRef
import ai.router.sdk.ProviderType
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

/**
 * A text-to-speech model at one provider, as listed by
 * `GET /v1/text-to-speech/models`, with the [voices] a request to it may name.
 *
 * [maxInputChars] is the longest text the model accepts per request (`0`
 * means unknown). [costPerMChars] is the price per million input characters;
 * `null` means unknown and `0.0` means free.
 */
@Serializable
public data class TextToSpeechModel(
    override val id: String,
    override val model: String,
    override val provider: String,
    @SerialName("provider_type") override val providerType: ProviderType,
    val voices: List<TextToSpeechVoice>,
    @SerialName("max_input_chars") val maxInputChars: Int = 0,
    @SerialName("cost_per_m_chars") val costPerMChars: Double? = null,
) : ModelRef

/**
 * A voice a model speaks in. Pass [id] as a request's voice.
 */
@Serializable
public data class TextToSpeechVoice(
    val id: String,
    val name: String,
)
