import { AiRouterClient, type ReasoningEffort } from "ai-router-sdk";

// ─── Server + model targets ───────────────────────────────────────────

export const SERVER_URL = "http://localhost:8787";
export const CHAT_MODEL = "claude-sonnet-5-5:cloud@anthropic";
export const CHAT_EFFORT: ReasoningEffort = "low";
export const EMBED_MODEL = "text-embedding-3-small:cloud@openai";
export const SEARCH_MODEL = "fast:cloud@exa";
export const CONTENTS_MODEL = "auto:cloud@exa";
export const TEXT_TO_SPEECH_MODEL = "eleven_flash_v2_5:cloud@elevenlabs";
export const SPEECH_TO_TEXT_MODEL = "scribe_v2:cloud@elevenlabs";

// ─── Shared client factory ────────────────────────────────────────────

export function newExampleClient(): AiRouterClient {
  return new AiRouterClient({
    baseUrl: SERVER_URL,
    fetch: (url, init) => fetch(url, { ...init, signal: AbortSignal.timeout(300_000) }),
  });
}
