package ai.router.sdk.examples

import ai.router.sdk.dsl.chatRequest
import ai.router.sdk.models.AiRouterException
import kotlinx.coroutines.runBlocking
import kotlin.test.assertTrue
import org.junit.jupiter.api.Assertions.assertThrows
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test

class ErrorHandlingExample {

    @Test
    @DisplayName("Error handling: non-2xx responses surface as AiRouterException")
    fun run() {
        val ex = assertThrows(AiRouterException::class.java) {
            runBlocking {
                newExampleClient().use { client ->
                    client.chat(chatRequest("does-not-exist:local@nonexistent") {
                        messages { user { text("hi") } }
                    })
                }
            }
        }

        assertTrue(ex.statusCode in 400..599, "expected a 4xx/5xx status, got ${ex.statusCode}")
        assertTrue(ex.apiError.type.isNotBlank(), "expected a non-blank error type")
        assertTrue(ex.apiError.message.isNotBlank(), "expected a non-blank error message")
    }
}
