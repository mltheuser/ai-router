import type { ChatMessage } from "./chatMessage";
import type { ChatSettings } from "./chatSettings";
import type { ResponseFormatWire } from "./responseFormat";

/** The body of `POST /v1/chat`. */
export interface ChatRequest extends Omit<ChatSettings, "response_format"> {
  messages: ChatMessage[];
  response_format?: ResponseFormatWire;
}

/** The request body for `messages` answered as `settings` say. */
export function chatRequest(settings: ChatSettings, messages: ChatMessage[]): ChatRequest {
  const { response_format: responseFormat, ...rest } = settings;
  if (responseFormat === undefined) {
    return { ...rest, messages };
  }
  return { ...rest, messages, response_format: responseFormat.toWire() };
}
