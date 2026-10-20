package ai.router.sdk.contents

import kotlinx.serialization.Serializable

/**
 * Page contents request: the pages to load.
 */
@Serializable
public data class ContentsRequest(
    val model: String,
    val urls: List<String>,
)
