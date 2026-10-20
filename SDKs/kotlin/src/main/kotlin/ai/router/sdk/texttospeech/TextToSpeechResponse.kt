package ai.router.sdk.texttospeech

import kotlinx.serialization.Serializable

/**
 * Text-to-speech response. [audio] is a base64-encoded WAV file of 16-bit
 * mono PCM, at a sample rate up to the provider.
 */
@Serializable
public data class TextToSpeechResponse(
    val model: String,
    val audio: String,
)
