package ai.router.sdk.texttospeech

import kotlinx.serialization.Serializable

/**
 * Text-to-speech request: [text] spoken in [voice], the id of one of the
 * voices listed in [model]'s entry.
 */
@Serializable
public data class TextToSpeechRequest(
    val model: String,
    val voice: String,
    val text: String,
)
