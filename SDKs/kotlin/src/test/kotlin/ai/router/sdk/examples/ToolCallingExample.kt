package ai.router.sdk.examples

import ai.router.sdk.dsl.chatRequest
import ai.router.sdk.models.decode
import ai.router.sdk.schema.Description
import kotlinx.coroutines.runBlocking
import kotlinx.serialization.Serializable
import org.junit.jupiter.api.Assumptions.assumeTrue
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
        newExampleClient().use { client ->
            // Initial request — the model should respond with a tool call.
            val first = client.chat(
                chatRequest(CHAT_MODEL) {
                    messages {
                        user { text("What's the weather in Berlin?") }
                    }
                    tools {
                        tool<WeatherParams>("get_weather", "Get current weather for a city")
                    }
                }
            )

            val toolCall = first.message.toolCalls?.firstOrNull()
            // Local models occasionally decline to call the tool; report a JUnit
            // skip rather than silently passing or flaking the test.
            assumeTrue(
                toolCall != null,
                "model produced no tool call (finish_reason=${first.finishReason})",
            )
            requireNotNull(toolCall)

            // Decoding the arguments validates the SDK's tool-call type surface.
            val params = toolCall.decode<WeatherParams>()
            assertTrue(params.city.isNotBlank(), "decoded tool call had blank city: $params")

            // Follow-up turn with the tool result attached.
            val followUp = client.chat(
                chatRequest(CHAT_MODEL) {
                    messages {
                        user { text("What's the weather in Berlin?") }
                        assistant { text("") } // assistant turn that issued the tool call
                        tool(callId = toolCall.id) {
                            text("""{"temp_celsius": 22, "condition": "sunny"}""")
                        }
                    }
                }
            )

            assertTrue(
                followUp.textContent.isNotBlank(),
                "expected a non-empty follow-up reply (finish_reason=${followUp.finishReason})",
            )
        }
    }
}
