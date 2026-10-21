import { z } from "zod";

/**
 * The JSON Schema of the values `schema` accepts, as declared: the SDK adds
 * nothing to it. It describes the schema's input, which is what `parse`
 * checks, so a value matching it parses.
 */
export function toJsonSchema(schema: z.ZodType): Record<string, unknown> {
  return z.toJSONSchema(schema, { io: "input" });
}
