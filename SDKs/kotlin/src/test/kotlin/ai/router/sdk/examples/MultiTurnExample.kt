package ai.router.sdk.examples

import ai.router.sdk.AiRouterClient
import ai.router.sdk.chat.ChatMessage
import ai.router.sdk.chat.ChatRole
import ai.router.sdk.chat.ChatSettings
import ai.router.sdk.chat.ContentPart
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test
import kotlin.test.assertTrue

class MultiTurnExample {

    @Test
    @DisplayName("Multi-turn conversation with sampling parameters")
    fun run() = runBlocking {
        AiRouterClient(SERVER_URL).use { client ->
            val settings = ChatSettings(CHAT_MODEL, CHAT_EFFORT, temperature = 0.7, maxTokens = 2048)
            val conversation = mutableListOf(
                ChatMessage(ChatRole.SYSTEM, listOf(ContentPart.Text("You are a helpful assistant."))),
                ChatMessage(ChatRole.USER, listOf(ContentPart.Text("My name is Alice."))),
            )

            // Each reply joins the conversation as received, then the next question.
            conversation += client.chat.send(settings, conversation).message
            conversation += ChatMessage(ChatRole.USER, listOf(ContentPart.Text("What's my name?")))
            val response = client.chat.send(settings, conversation)

            println("Answer: ${response.message.text}")
            assertTrue(
                response.message.text.isNotBlank(),
                "expected non-empty text content (finish_reason=${response.finishReason})",
            )
        }
    }
}
