import type { ChatMessage } from "./chatMessage";

/** The answer to a chat request: the model's next message. */
export interface ChatResponse {
  model: string;
  message: ChatMessage;
  finish_reason: FinishReason;
  usage: ChatUsage;
}

/** Why the model stopped generating. Providers may return values beyond the named ones. */
export type FinishReason =
  | "stop"
  | "tool_calls"
  | "length"
  | "content_filter"
  | "error"
  | (string & {});

/** The tokens a request consumed. */
export interface ChatUsage {
  /** The whole prompt, cached tokens included. */
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
  reasoning_tokens: number;
  cache_read_tokens: number;
}
