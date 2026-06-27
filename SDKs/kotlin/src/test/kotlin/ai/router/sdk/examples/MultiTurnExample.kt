package ai.router.sdk.examples

import ai.router.sdk.dsl.chatRequest
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test
import kotlin.test.assertTrue

class MultiTurnExample {

    @Test
    @DisplayName("Multi-turn conversation with sampling parameters")
    fun run() = runBlocking {
        newExampleClient().use { client ->
            val response = client.chat(
                chatRequest(CHAT_MODEL) {
                    messages {
                        system { text("You are a helpful assistant.") }
                        user { text("My name is Alice.") }
                        assistant { text("Hello Alice! How can I help you?") }
                        user { text("What's my name?") }
                    }
                    temperature(0.7)
                    maxTokens(512)
                }
            )

            assertTrue(
                response.textContent.isNotBlank(),
                "expected non-empty text content (finish_reason=${response.choices.finishReason})",
            )
        }
    }
}
