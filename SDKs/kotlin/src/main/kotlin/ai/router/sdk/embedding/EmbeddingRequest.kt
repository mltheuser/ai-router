package ai.router.sdk.embedding

import kotlinx.serialization.Serializable

/** The body of `POST /v1/embedding`: one vector per text of [input]. [dimensions] is not supported by every model. */
@Serializable
public data class EmbeddingRequest(
    val model: String,
    val input: List<String>,
    val dimensions: Int? = null,
)
