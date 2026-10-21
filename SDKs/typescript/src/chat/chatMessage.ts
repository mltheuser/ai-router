import type { ToolCall } from "./toolCall";

/**
 * A message of a conversation, which is a `ChatMessage[]`. A caller writes
 * system, user and tool messages; assistant messages are the model's replies,
 * added to the conversation as received.
 */
export interface ChatMessage {
  role: ChatRole;
  content: ContentPart[];
  /** The model's reasoning trace, on an assistant message. */
  reasoning_content?: string;
  /** The tools the model calls, on an assistant message; each needs a tool message answering it. */
  tool_calls?: ToolCall[];
  /** The {@link ToolCall.id} a tool message answers. */
  tool_call_id?: string;
}

/** The author of a message: the four roles the router accepts. */
export type ChatRole = "system" | "user" | "assistant" | "tool";

/** One piece of a message. */
export type ContentPart = TextPart | ImagePart;

export interface TextPart {
  type: "text";
  /** Absent on an empty text part. */
  text?: string;
}

/** An image inline as base64 without a data-URI prefix. */
export interface ImagePart {
  type: "image";
  /** E.g. `image/png`, `image/jpeg`, `image/webp` or `image/gif`. */
  mime_type: string;
  base64_data: string;
}

/** The message's text parts, concatenated. */
export function messageText(message: ChatMessage): string {
  let text = "";
  for (const part of message.content) {
    if (part.type === "text") {
      text += part.text ?? "";
    }
  }
  return text;
}
