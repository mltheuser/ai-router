package ai.router.sdk.examples

import ai.router.sdk.models.Capability
import ai.router.sdk.models.ProviderType
import kotlinx.coroutines.runBlocking
import org.junit.jupiter.api.DisplayName
import org.junit.jupiter.api.Test
import kotlin.test.assertEquals
import kotlin.test.assertTrue

class ListModelsExample {

    @Test
    @DisplayName("List models: fetch the catalog, with and without filters")
    fun run() = runBlocking {
        newExampleClient().use { client ->
            // The full catalog: every model the router can serve right now.
            val all = client.listModels()

            assertEquals("list", all.`object`)
            assertTrue(all.data.isNotEmpty(), "expected a configured server to expose at least one model")

            assertTrue(
                all.data.all { it.model.startsWith("${it.id}:") && it.model.endsWith("@${it.provider}") },
                "expected every entry's model string to qualify its id with tag and provider",
            )

            // Narrow the catalog with optional filters: provider type,
            // capability (chat/embed), and a case-insensitive id search.
            val localChat = client.listModels(
                type = ProviderType.LOCAL,
                capability = Capability.CHAT,
            )

            assertTrue(
                localChat.data.all { it.providerType == ProviderType.LOCAL && it.hasCapability(Capability.CHAT) },
                "expected every filtered model to be a local chat model",
            )

            // Search matches model ids case-insensitively; reuse a fragment
            // of an id we know is in the catalog.
            val fragment = all.data.first().id.take(4)
            val searched = client.listModels(search = fragment)

            assertTrue(
                searched.data.isNotEmpty() && searched.data.all { it.id.contains(fragment, ignoreCase = true) },
                "expected every searched model id to contain \"$fragment\"",
            )
        }
    }
}
