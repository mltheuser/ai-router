package ai.router.sdk.speechtotext

import kotlinx.serialization.Serializable

/**
 * Speech-to-text request. [audio] is a base64-encoded WAV file of 16-bit mono
 * PCM, at any sample rate: the format text-to-speech returns, so a
 * text-to-speech response's `audio` can be passed as is.
 */
@Serializable
public data class SpeechToTextRequest(
    val model: String,
    val audio: String,
)
