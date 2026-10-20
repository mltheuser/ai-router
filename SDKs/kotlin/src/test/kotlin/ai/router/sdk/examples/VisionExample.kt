package ai.router.sdk.examples

import ai.router.sdk.AiRouterClient
import ai.router.sdk.chat.ChatMessage
import ai.router.sdk.chat.ChatRole
import ai.router.sdk.chat.ChatSettings
import ai.router.sdk.chat.ContentPart
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

        AiRouterClient(SERVER_URL).use { client ->
            val response = client.chat.send(
                ChatSettings(CHAT_MODEL, CHAT_EFFORT),
                listOf(
                    ChatMessage(
                        ChatRole.USER,
                        listOf(
                            ContentPart.Text("Briefly describe this image."),
                            ContentPart.Image(mimeType = "image/png", base64Data = imageBase64),
                        ),
                    ),
                ),
            )

            assertTrue(
                response.message.text.isNotBlank(),
                "expected non-empty text content (finish_reason=${response.finishReason})",
            )
        }
    }
}
