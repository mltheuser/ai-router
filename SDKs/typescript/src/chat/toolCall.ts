import type { z } from "zod";

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
 * The call's arguments, parsed by the schema of the tool's parameters: as
 * strict as that schema, e.g. a `z.strictObject` rejects a key it does not
 * declare. Throws a `ZodError` on arguments that do not match.
 */
export function decodeToolCall<S extends z.ZodType>(call: ToolCall, schema: S): z.output<S> {
  return schema.parse(call.function.arguments);
}
