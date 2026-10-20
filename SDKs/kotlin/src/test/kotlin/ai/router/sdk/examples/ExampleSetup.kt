package ai.router.sdk.examples

import ai.router.sdk.chat.ReasoningEffort

// ─── Server + model targets ───────────────────────────────────────────

internal const val SERVER_URL = "http://localhost:8787"
internal const val CHAT_MODEL = "claude-sonnet-5-5:cloud@anthropic"

internal val CHAT_EFFORT = ReasoningEffort.LOW
internal const val EMBED_MODEL = "text-embedding-3-small:cloud@openai"
internal const val SEARCH_MODEL = "fast:cloud@exa"
internal const val CONTENTS_MODEL = "auto:cloud@exa"
internal const val TEXT_TO_SPEECH_MODEL = "eleven_flash_v2_5:cloud@elevenlabs"
internal const val SPEECH_TO_TEXT_MODEL = "scribe_v2:cloud@elevenlabs"
