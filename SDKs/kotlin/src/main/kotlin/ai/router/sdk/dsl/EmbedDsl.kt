package ai.router.sdk.dsl

import ai.router.sdk.models.EmbedRequest

/**
 * Entry point for building an [EmbedRequest] via DSL.
 *
 * ```kotlin
 * val request = embedRequest("nomic-embed-text:local") {
 *     batch("First document to embed")
 *     batch("Second document to embed")
 *     dimensions(768)
 * }
 * ```
 */
fun embedRequest(model: String, block: EmbedRequestBuilder.() -> Unit): EmbedRequest {
    return EmbedRequestBuilder(model).apply(block).build()
}

@ChatDsl
class EmbedRequestBuilder(private val model: String) {
    private val inputs = mutableListOf<String>()
    private var dimensions: Int? = null

    /**
     * Add a text to embed as a separate batch entry.
     * Each [batch] call corresponds to one embedding in the response.
     */
    fun batch(text: String) {
        inputs.add(text)
    }

    fun dimensions(value: Int) {
        dimensions = value
    }

    internal fun build(): EmbedRequest {
        require(inputs.isNotEmpty()) { "At least one batch() call is required" }
        return EmbedRequest(
            model = model,
            input = inputs.toList(),
            dimensions = dimensions,
        )
    }
}
