package ai.router.sdk.examples

import ai.router.sdk.dsl.structuredChatRequest
import kotlinx.coroutines.runBlocking
import kotlinx.serialization.Serializable
import kotlin.test.assertTrue
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test

class StructuredOutputExample {

    @Serializable
    data class WeatherInfo(
        val city: String,
        val tempCelsius: Double,
        val condition: String,
    )

    @Test
    @DisplayName("Structured output: deserialize into a @Serializable class")
    fun run() = runBlocking {
        newExampleClient().use { client ->
            val request = structuredChatRequest<WeatherInfo>(CHAT_MODEL) {
                messages {
                    system { text("Extract weather information from the text.") }
                    user { text("It's 22°C and sunny in Berlin today.") }
                }
            }

            val weather: WeatherInfo = client.chat(request)

            assertTrue(weather.city.isNotBlank(), "expected non-empty city (got $weather)")
        }
    }
}
