import type { ModelRef } from "../models";

/** An optional ability of a chat model beyond plain text chat. */
export type ChatFeature = "tools" | "vision" | "reasoning" | "structured_output";

/**
 * A chat model at one provider, as listed by `GET /v1/chat/models`. Cost
 * fields are per million tokens; absent means unknown and `0` means free.
 */
export interface ChatModel extends ModelRef {
  features?: ChatFeature[];
  context_window?: number;
  cost_per_m_input?: number;
  cost_per_m_output?: number;
  /** Set for local models. */
  size_bytes?: number;
}

/** The author of a chat message. */
export type ChatRole = "system" | "user" | "assistant" | "tool";

/**
 * How much reasoning the model performs; `none` turns it off. A model that
 * cannot honour the level rejects the request.
 */
export type ReasoningEffort = "none" | "low" | "medium" | "high";

/** Why the model stopped generating. Providers may return values beyond the named ones. */
export type FinishReason =
  | "stop"
  | "tool_calls"
  | "length"
  | "content_filter"
  | "error"
  | (string & {});

/** Discriminates the parts of a multimodal message. */
export type ContentPartType = "text" | "image";

/** One piece of a multimodal message. */
export interface ContentPart {
  type: ContentPartType;
  /** For type `text`. */
  text?: string;
  /** For type `image`, e.g. `image/png`, `image/jpeg`, `image/webp` or `image/gif`. */
  mime_type?: string;
  /** For type `image`: raw base64, without a data-URI prefix. */
  base64_data?: string;
}

/** A plain JSON Schema object. */
export type JsonSchema = Record<string, unknown>;

/** A tool the model may call, its arguments described by `parameters`. */
export interface ToolDefinition {
  name: string;
  description?: string;
  parameters?: JsonSchema;
}

/** A tool invocation requested by the assistant. `id` matches a tool result to its call. */
export interface ToolCall {
  id: string;
  function: ToolCallFunction;
}

/** The function invocation inside a {@link ToolCall}. */
export interface ToolCallFunction {
  name: string;
  arguments: Record<string, unknown>;
}

/** The named schema of a {@link ResponseFormat}. */
export interface JsonSchemaSpec {
  name: string;
  description?: string;
  schema?: JsonSchema;
}

/** Constrains the response's text to JSON matching a schema (structured output). */
export interface ResponseFormat {
  type: "json_schema";
  json_schema?: JsonSchemaSpec;
}

/**
 * A message of the conversation. An assistant message may carry
 * `reasoning_content` and `tool_calls`; a tool result carries the
 * `tool_call_id` it answers.
 */
export interface ChatMessage {
  role: ChatRole;
  content: ContentPart[];
  reasoning_content?: string;
  tool_calls?: ToolCall[];
  tool_call_id?: string;
}

/** The body of `POST /v1/chat`. */
export interface ChatRequest {
  model: string;
  messages: ChatMessage[];
  frequency_penalty?: number;
  max_tokens?: number;
  presence_penalty?: number;
  temperature?: number;
  top_p?: number;
  response_format?: ResponseFormat;
  reasoning_effort: ReasoningEffort;
  tools?: ToolDefinition[];
}

/** The tokens a request consumed. */
export interface ChatUsage {
  /** The whole prompt, cached tokens included. */
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
  reasoning_tokens: number;
  cache_read_tokens: number;
}

/** The answer to a chat request: the model's next message. */
export interface ChatResponse {
  model: string;
  message: ChatMessage;
  finish_reason: FinishReason;
  usage: ChatUsage;
}

/** The text of the response's message: its text parts, concatenated. */
export function chatText(response: ChatResponse): string {
  return response.message.content
    .filter((part) => part.type === "text")
    .map((part) => part.text ?? "")
    .join("");
}

/** The model's reasoning trace, or `""` if it returned none. */
export function chatReasoning(response: ChatResponse): string {
  return response.message.reasoning_content ?? "";
}

/**
 * The response's text parsed as JSON, for a request with a `response_format`.
 * The result is cast to `T`, not validated against the schema.
 */
export function parseStructuredOutput<T>(response: ChatResponse): T {
  return JSON.parse(chatText(response)) as T;
}
