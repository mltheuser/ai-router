package ai.router.sdk.examples

import ai.router.sdk.speechtotext.SpeechToTextRequest
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test
import kotlin.test.assertTrue

class SpeechToTextExample {

    @Test
    @DisplayName("Speech to text: transcribe a WAV file")
    fun run() = runBlocking {
        // A WAV file of 16-bit mono PCM at 24 kHz, made by our text to speech.
        // It says "The quick brown fox jumps over the lazy dog."
        val wav = checkNotNull(javaClass.getResourceAsStream("/sentence.wav")) {
            "sentence.wav missing from src/test/resources"
        }.use { it.readBytes() }

        newExampleClient().use { client ->
            val response = client.speechToText.send(SpeechToTextRequest.fromWav(SPEECH_TO_TEXT_MODEL, wav))

            println("Transcript: ${response.text}")
            assertTrue(response.text.isNotBlank(), "expected a transcript")
        }
    }
}
