import { decodeJsonWith, toJsonSchema } from "../schema/schema";
import type { StandardSchemaWithJson } from "../schema/standardSchema";
import { messageText } from "./chatMessage";
import type { ChatResponse } from "./chatResponse";

export interface ResponseFormatOptions {
  name: string;
  description?: string;
}

/**
 * Asks the model to answer with JSON of `schema`'s JSON Schema, sent as the
 * request's `response_format` as declared. A response to such a request is an
 * ordinary response, which {@link decode} reads.
 */
export class ResponseFormat<T> {
  readonly schema: StandardSchemaWithJson<T>;
  readonly name: string;
  readonly description: string | undefined;

  constructor(schema: StandardSchemaWithJson<T>, options: ResponseFormatOptions) {
    this.schema = schema;
    this.name = options.name;
    this.description = options.description;
  }

  /**
   * The text of `response`'s message, parsed and validated by the schema.
   * Throws `DecodeError` if it does not match, e.g. when the model called
   * tools instead of answering, or was cut off.
   */
  decode(response: ChatResponse): T {
    return decodeJsonWith(this.schema, messageText(response.message));
  }

  /** @internal The wire's `response_format`. */
  toWire(): ResponseFormatWire {
    let schema: Record<string, unknown>;
    try {
      schema = toJsonSchema(this.schema);
    } catch (cause) {
      throw new TypeError(`response format '${this.name}' has no JSON Schema`, { cause });
    }
    const wire: ResponseFormatWire = {
      type: "json_schema",
      json_schema: { name: this.name, schema },
    };
    if (this.description !== undefined) {
      wire.json_schema.description = this.description;
    }
    return wire;
  }
}

/** @internal */
export interface ResponseFormatWire {
  type: "json_schema";
  json_schema: { name: string; description?: string; schema: Record<string, unknown> };
}
