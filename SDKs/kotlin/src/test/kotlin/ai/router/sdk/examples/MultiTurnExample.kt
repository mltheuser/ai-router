package ai.router.sdk.examples

import ai.router.sdk.AiRouterClient
import ai.router.sdk.chat.ChatMessage
import ai.router.sdk.chat.ContentPart
import ai.router.sdk.chat.ContentPartType
import ai.router.sdk.chat.chatRequest
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test
import kotlin.test.assertTrue

class MultiTurnExample {

    @Test
    @DisplayName("Multi-turn conversation with sampling parameters")
    fun run() = runBlocking {
        AiRouterClient(SERVER_URL).use { client ->
            val request = chatRequest(CHAT_MODEL, CHAT_EFFORT) {
                messages {
                    system { text("You are a helpful assistant.") }
                    user { text("My name is Alice.") }
                }
                temperature(0.7)
                maxTokens(2048)
            }

            // The conversation: each reply joins it as received, then the next question.
            val history = request.messages.toMutableList()
            history += client.chat.send(request).message
            history += ChatMessage(
                role = "user",
                content = listOf(ContentPart(type = ContentPartType.TEXT, text = "What's my name?")),
            )
            val response = client.chat.send(request.copy(messages = history))

            println("Answer: ${response.textContent}")
            assertTrue(
                response.textContent.isNotBlank(),
                "expected non-empty text content (finish_reason=${response.finishReason})",
            )
        }
    }
}
