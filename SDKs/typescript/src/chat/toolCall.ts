import { decodeWith } from "../schema/schema";
import type { StandardSchemaWithJson } from "../schema/standardSchema";

/** A tool invocation requested by the model; `id` matches a tool result to its call. */
export interface ToolCall {
  id: string;
  function: ToolCallFunction;
}

export interface ToolCallFunction {
  name: string;
  arguments: Record<string, unknown>;
}

/**
 * The call's arguments, validated by the schema of the tool's parameters:
 * as strict as that schema, e.g. a `z.strictObject` rejects a key it does not
 * declare. Throws `DecodeError` on arguments that do not match.
 */
export function decodeToolCall<T>(call: ToolCall, schema: StandardSchemaWithJson<T>): T {
  return decodeWith(schema, call.function.arguments);
}
