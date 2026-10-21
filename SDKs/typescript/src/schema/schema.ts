import { DecodeError } from "./decodeError";
import type { StandardSchemaWithJson } from "./standardSchema";

/**
 * The JSON Schema of the values `schema` accepts, as declared: the SDK adds
 * nothing to it. It is the schema the library's validation checks, so a
 * value matching it decodes.
 */
export function toJsonSchema(schema: StandardSchemaWithJson): Record<string, unknown> {
  return schema["~standard"].jsonSchema.input({ target: "draft-2020-12" });
}

/** `value` validated by `schema`; throws {@link DecodeError} if it does not match. */
export function decodeWith<T>(schema: StandardSchemaWithJson<T>, value: unknown): T {
  const result = schema["~standard"].validate(value);
  if (result instanceof Promise) {
    throw new TypeError(
      "schemas that validate asynchronously are not supported; use a synchronous schema",
    );
  }
  if (result.issues !== undefined) {
    throw new DecodeError(result.issues);
  }
  return result.value;
}

/** `text` parsed as JSON and validated by `schema`; throws {@link DecodeError} if it is not JSON or does not match. */
export function decodeJsonWith<T>(schema: StandardSchemaWithJson<T>, text: string): T {
  let value: unknown;
  try {
    value = JSON.parse(text);
  } catch (cause) {
    throw new DecodeError([], { cause });
  }
  return decodeWith(schema, value);
}
