package ai.router.sdk.examples

import ai.router.sdk.dsl.chatRequest
import ai.router.sdk.models.ReasoningEffort
import kotlinx.coroutines.runBlocking
import kotlin.test.assertTrue
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test

class ReasoningExample {

    @Test
    @DisplayName("Reasoning: separate reasoning trace from the final answer")
    fun run() = runBlocking {
        newExampleClient().use { client ->
            val response = client.chat(chatRequest(CHAT_MODEL) {
                messages {
                    user { text("Prove that √2 is irrational.") }
                }
                reasoningEffort(ReasoningEffort.HIGH)
            })

            assertTrue(
                response.textContent.isNotBlank(),
                "expected non-empty answer (finish_reason=${response.choices.finishReason})",
            )
            assertTrue(
                !response.choices.message.reasoningContent.isNullOrBlank(),
                "expected reasoning trace from a reasoning-capable model",
            )
        }
    }
}
