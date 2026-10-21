import type { ResponseFormat } from "./responseFormat";
import type { ToolDefinition } from "./toolDefinition";

/** Everything of a chat request but its messages: which model answers, and how. */
export interface ChatSettings {
  model: string;
  reasoning_effort: ReasoningEffort;
  tools?: ToolDefinition[];
  response_format?: ResponseFormat;
  temperature?: number;
  top_p?: number;
  max_tokens?: number;
  frequency_penalty?: number;
  presence_penalty?: number;
}

/**
 * How much reasoning the model performs; `none` turns it off. A model that
 * cannot honour the level rejects the request.
 */
export type ReasoningEffort = "none" | "low" | "medium" | "high";
