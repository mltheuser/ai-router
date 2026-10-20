package ai.router.sdk.examples

import ai.router.sdk.AiRouterClient
import ai.router.sdk.chat.ChatMessage
import ai.router.sdk.chat.ChatRole
import ai.router.sdk.chat.ChatSettings
import ai.router.sdk.chat.ContentPart
import ai.router.sdk.chat.ReasoningEffort
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test
import kotlin.test.assertTrue

class ReasoningExample {

    // A novel puzzle: models with adaptive thinking answer famous problems
    // without thinking, so a canonical one shows no trace.
    private val puzzle = "Three friends - Ana, Ben and Cy - each have a different pet (cat, dog, fish) " +
        "and live in houses 1, 2 and 3. Ana is not in house 1. The dog owner is in " +
        "house 2. Ben owns the fish. Cy is not in house 3. Who owns the cat, and in " +
        "which house?"

    @Test
    @DisplayName("Reasoning: separate reasoning trace from the final answer")
    fun run() = runBlocking {
        AiRouterClient(SERVER_URL).use { client ->
            val response = client.chat.send(
                ChatSettings(CHAT_MODEL, ReasoningEffort.HIGH),
                listOf(ChatMessage(ChatRole.USER, listOf(ContentPart.Text(puzzle)))),
            )

            assertTrue(
                response.message.text.isNotBlank(),
                "expected non-empty answer (finish_reason=${response.finishReason})",
            )
            assertTrue(
                !response.message.reasoningContent.isNullOrBlank(),
                "expected reasoning trace from a reasoning-capable model",
            )
        }
    }
}
