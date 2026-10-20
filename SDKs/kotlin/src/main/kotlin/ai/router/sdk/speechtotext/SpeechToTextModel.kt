package ai.router.sdk.speechtotext

import ai.router.sdk.ModelRef
import ai.router.sdk.ProviderType
import kotlinx.serialization.SerialName
import kotlinx.serialization.Serializable

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
