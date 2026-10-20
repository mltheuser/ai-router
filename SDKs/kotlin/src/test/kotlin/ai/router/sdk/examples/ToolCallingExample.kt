package ai.router.sdk.examples

import ai.router.sdk.AiRouterClient
import ai.router.sdk.chat.ChatMessage
import ai.router.sdk.chat.ContentPart
import ai.router.sdk.chat.ContentPartType
import ai.router.sdk.chat.chatRequest
import ai.router.sdk.chat.decode
import ai.router.sdk.schema.Description
import kotlinx.coroutines.runBlocking
import kotlinx.serialization.Serializable
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test
import kotlin.test.assertTrue

class ToolCallingExample {

    @Serializable
    data class WeatherParams(
        @Description("The city to check weather for") val city: String,
        val unit: String = "celsius",
    )

    @Test
    @DisplayName("Tool calling: define a tool, decode the call, send the result back")
    fun run() = runBlocking {
        AiRouterClient(SERVER_URL).use { client ->
            val request = chatRequest(CHAT_MODEL, CHAT_EFFORT) {
                messages {
                    user { text("What's the weather in Berlin?") }
                }
                tools {
                    tool<WeatherParams>("get_weather", "Get current weather for a city")
                }
            }

            // The model should answer with a call of the tool.
            val first = client.chat.send(request)
            val toolCall = checkNotNull(first.message.toolCalls?.firstOrNull()) {
                "model produced no tool call (finish_reason=${first.finishReason})"
            }
            val params = toolCall.decode<WeatherParams>()
            assertTrue(params.city.isNotBlank(), "decoded tool call had blank city: $params")

            // The conversation continues with the model's reply as received, which
            // carries the call, and the tool's result for that call.
            val history = request.messages.toMutableList()
            history += first.message
            history += ChatMessage(
                role = "tool",
                content = listOf(
                    ContentPart(type = ContentPartType.TEXT, text = """{"temp_celsius": 22, "condition": "sunny"}"""),
                ),
                toolCallId = toolCall.id,
            )
            val followUp = client.chat.send(request.copy(messages = history))

            println("Answer: ${followUp.textContent}")
            assertTrue(
                followUp.textContent.isNotBlank(),
                "expected a non-empty follow-up reply (finish_reason=${followUp.finishReason})",
            )
        }
    }
}
