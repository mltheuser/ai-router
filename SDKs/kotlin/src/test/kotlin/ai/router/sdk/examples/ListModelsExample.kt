package ai.router.sdk.examples

import ai.router.sdk.ProviderType
import ai.router.sdk.chat.ChatFeature
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test
import kotlin.test.assertTrue

class ListModelsExample {

    @Test
    @DisplayName("List models: fetch each use case's models, with and without filters")
    fun run() = runBlocking {
        newExampleClient().use { client ->
            // Every use case lists its own models, each with metadata of its
            // own: chat models report their features, for example.
            val chatModels = client.chat.listModels()
            val embedModels = client.embedding.listModels()
            val searchModels = client.search.listModels()
            val contentsModels = client.contents.listModels()

            assertTrue(chatModels.data.isNotEmpty(), "expected a configured server to expose at least one chat model")
            assertTrue(
                (chatModels.data + embedModels.data + searchModels.data + contentsModels.data).all {
                    it.model.startsWith("${it.id}:") && it.model.endsWith("@${it.provider}")
                },
                "expected every entry's model string to qualify its id with tag and provider",
            )

            // Pick a model by what it can do.
            val toolUser = chatModels.data.firstOrNull { it.has(ChatFeature.TOOLS, ChatFeature.VISION) }
            println("A chat model with tools and vision: ${toolUser?.model}")

            // Narrow a listing by provider type and a case-insensitive id search.
            val fragment = chatModels.data.first().id.take(4)
            val narrowed = client.chat.listModels(type = chatModels.data.first().providerType, search = fragment)

            assertTrue(
                narrowed.data.isNotEmpty() && narrowed.data.all { it.id.contains(fragment, ignoreCase = true) },
                "expected every searched model id to contain \"$fragment\"",
            )

            val local = client.embedding.listModels(type = ProviderType.LOCAL)
            assertTrue(local.data.all { it.providerType == ProviderType.LOCAL }, "expected only local models")
        }
    }
}
