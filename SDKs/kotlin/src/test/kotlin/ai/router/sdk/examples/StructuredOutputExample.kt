package ai.router.sdk.examples

import ai.router.sdk.AiRouterClient
import ai.router.sdk.chat.ChatMessage
import ai.router.sdk.chat.ChatRole
import ai.router.sdk.chat.ChatSettings
import ai.router.sdk.chat.ContentPart
import ai.router.sdk.chat.ResponseFormat
import ai.router.sdk.schema.NoAdditionalProperties
import kotlinx.coroutines.runBlocking
import kotlinx.serialization.Serializable
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test
import kotlin.test.assertTrue

class StructuredOutputExample {

    // Anthropic and OpenAI accept only objects that rule out further properties.
    @Serializable
    @NoAdditionalProperties
    data class WeatherInfo(
        val city: String,
        val tempCelsius: Double,
        val condition: String,
    )

    @Test
    @DisplayName("Structured output: decode the answer into a @Serializable class")
    fun run() = runBlocking {
        AiRouterClient(SERVER_URL).use { client ->
            val weatherFormat = ResponseFormat(WeatherInfo.serializer())
            val instruction = "Extract weather information from the text."
            val response = client.chat.send(
                ChatSettings(CHAT_MODEL, CHAT_EFFORT, responseFormat = weatherFormat),
                listOf(
                    ChatMessage(ChatRole.SYSTEM, listOf(ContentPart.Text(instruction))),
                    ChatMessage(ChatRole.USER, listOf(ContentPart.Text("It's 22°C and sunny in Berlin today."))),
                ),
            )

            val weather: WeatherInfo = weatherFormat.decode(response)

            assertTrue(weather.city.isNotBlank(), "expected non-empty city (got $weather)")
        }
    }
}
