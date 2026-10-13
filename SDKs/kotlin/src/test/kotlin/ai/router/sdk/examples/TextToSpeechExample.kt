package ai.router.sdk.examples

import ai.router.sdk.texttospeech.TextToSpeechRequest
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test
import kotlin.test.assertTrue

class TextToSpeechExample {

    @Test
    @DisplayName("Text to speech: speak a sentence in one of the model's voices")
    fun run() = runBlocking {
        newExampleClient().use { client ->
            // The model's entry lists the voices it speaks in.
            val entry = client.textToSpeech.listModels().data.first { it.model == TEXT_TO_SPEECH_MODEL }
            val voice = entry.voices.first()

            val response = client.textToSpeech.send(
                TextToSpeechRequest(
                    model = TEXT_TO_SPEECH_MODEL,
                    voice = voice.id,
                    text = "Hello from the ai-router Kotlin SDK.",
                )
            )

            val wav = response.wav()
            println("${voice.name}: ${wav.size} bytes of WAV audio")
            assertTrue(wav.isNotEmpty(), "expected audio")
        }
    }
}
