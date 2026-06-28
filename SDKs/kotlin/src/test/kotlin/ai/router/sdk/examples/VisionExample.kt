package ai.router.sdk.examples

import ai.router.sdk.dsl.chatRequest
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test
import java.util.Base64
import kotlin.test.assertTrue

class VisionExample {

    @Test
    @DisplayName("Vision: send an image alongside a text prompt")
    fun run() = runBlocking {
        val imageBytes = checkNotNull(javaClass.getResourceAsStream("/apple.png")) {
            "apple.png missing from src/test/resources"
        }.use { it.readBytes() }
        val imageBase64 = Base64.getEncoder().encodeToString(imageBytes)

        newExampleClient().use { client ->
            val response = client.chat(
                chatRequest(CHAT_MODEL) {
                    messages {
                        user {
                            text("Briefly describe this image.")
                            image(mimeType = "image/png", base64Data = imageBase64)
                        }
                    }
                }
            )

            assertTrue(
                response.textContent.isNotBlank(),
                "expected non-empty text content (finish_reason=${response.finishReason})",
            )
        }
    }
}
