package ai.router.sdk.speechtotext

import kotlinx.serialization.Serializable

/**
 * Speech-to-text response: [text] is the transcript of the whole audio.
 */
@Serializable
public data class SpeechToTextResponse(
    val model: String,
    val text: String,
)
