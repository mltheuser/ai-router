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

class QuickStartExample {

    @Test
    @DisplayName("Quick start: send a chat request and read the response")
    fun run() = runBlocking {
        AiRouterClient(SERVER_URL).use { client ->
            val response = client.chat.send(
                ChatSettings(CHAT_MODEL, CHAT_EFFORT),
                listOf(
                    ChatMessage(ChatRole.SYSTEM, listOf(ContentPart.Text("You are a helpful assistant."))),
                    ChatMessage(ChatRole.USER, listOf(ContentPart.Text("What is the capital of France?"))),
                ),
            )

            assertTrue(
                response.message.text.isNotBlank(),
                "expected non-empty text content (finish_reason=${response.finishReason})",
            )
        }
    }
}
