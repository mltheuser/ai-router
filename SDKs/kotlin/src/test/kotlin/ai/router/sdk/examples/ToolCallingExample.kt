package ai.router.sdk.examples

import ai.router.sdk.AiRouterClient
import ai.router.sdk.chat.ChatMessage
import ai.router.sdk.chat.ChatRole
import ai.router.sdk.chat.ChatSettings
import ai.router.sdk.chat.ContentPart
import ai.router.sdk.chat.ToolDefinition
import ai.router.sdk.schema.Description
import ai.router.sdk.schema.SchemaGenerator
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
    @DisplayName("Tool calling: define a tool, decode the calls, send the results back")
    fun run() = runBlocking {
        AiRouterClient(SERVER_URL).use { client ->
            val getWeather = ToolDefinition(
                name = "get_weather",
                description = "Get current weather for a city",
                parameters = SchemaGenerator.generate(WeatherParams.serializer()),
            )
            val settings = ChatSettings(CHAT_MODEL, CHAT_EFFORT, tools = listOf(getWeather))
            val conversation = mutableListOf(
                ChatMessage(ChatRole.USER, listOf(ContentPart.Text("What's the weather in Berlin?"))),
            )

            // The reply carries the model's calls; it joins the conversation as received.
            val reply = client.chat.send(settings, conversation).message
            conversation += reply
            assertTrue(reply.toolCalls.isNotEmpty(), "model produced no tool call")

            // Every call needs its result before the conversation can continue.
            for (call in reply.toolCalls) {
                val params = call.decode(WeatherParams.serializer())
                assertTrue(params.city.isNotBlank(), "decoded tool call had blank city: $params")
                conversation += ChatMessage(
                    role = ChatRole.TOOL,
                    content = listOf(ContentPart.Text("""{"temp_celsius": 22, "condition": "sunny"}""")),
                    toolCallId = call.id,
                )
            }
            val followUp = client.chat.send(settings, conversation)

            println("Answer: ${followUp.message.text}")
            assertTrue(
                followUp.message.text.isNotBlank(),
                "expected a non-empty follow-up reply (finish_reason=${followUp.finishReason})",
            )
        }
    }
}
