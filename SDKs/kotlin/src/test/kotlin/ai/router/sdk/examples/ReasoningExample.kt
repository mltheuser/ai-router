package ai.router.sdk.examples

import ai.router.sdk.chat.ReasoningEffort
import ai.router.sdk.chat.chatRequest
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test
import kotlin.test.assertTrue

class ReasoningExample {

    @Test
    @DisplayName("Reasoning: separate reasoning trace from the final answer")
    fun run() = runBlocking {
        newExampleClient().use { client ->
            val response = client.chat.send(
                chatRequest(CHAT_MODEL, ReasoningEffort.HIGH) {
                    messages {
                        // A novel puzzle: models with adaptive thinking answer famous
                        // problems without thinking, so a canonical one shows no trace.
                        user {
                            text(
                                "Three friends - Ana, Ben and Cy - each have a different pet (cat, dog, fish) " +
                                    "and live in houses 1, 2 and 3. Ana is not in house 1. The dog owner is in " +
                                    "house 2. Ben owns the fish. Cy is not in house 3. Who owns the cat, and in " +
                                    "which house?"
                            )
                        }
                    }
                }
            )

            assertTrue(
                response.textContent.isNotBlank(),
                "expected non-empty answer (finish_reason=${response.finishReason})",
            )
            assertTrue(
                !response.message.reasoningContent.isNullOrBlank(),
                "expected reasoning trace from a reasoning-capable model",
            )
        }
    }
}
