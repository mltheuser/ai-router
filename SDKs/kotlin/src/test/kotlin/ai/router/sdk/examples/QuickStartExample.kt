package ai.router.sdk.examples

import ai.router.sdk.dsl.chatRequest
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test
import kotlin.test.assertTrue

class QuickStartExample {

    @Test
    @DisplayName("Quick start: send a chat request and read the response")
    fun run() = runBlocking {
        newExampleClient().use { client ->
            val response = client.chat(
                chatRequest(CHAT_MODEL) {
                    messages {
                        system { text("You are a helpful assistant.") }
                        user { text("What is the capital of France?") }
                    }
                }
            )

            assertTrue(
                response.textContent.isNotBlank(),
                "expected non-empty text content (finish_reason=${response.choices.finishReason})",
            )
        }
    }
}
