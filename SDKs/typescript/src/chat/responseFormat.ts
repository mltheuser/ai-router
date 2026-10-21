import type { z } from "zod";
import { toJsonSchema } from "../schema/jsonSchema";
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
export class ResponseFormat<S extends z.ZodType = z.ZodType> {
  readonly schema: S;
  readonly name: string;
  readonly description: string | undefined;

  constructor(schema: S, options: ResponseFormatOptions) {
    this.schema = schema;
    this.name = options.name;
    this.description = options.description;
  }

  /**
   * The text of `response`'s message, parsed as JSON and by the schema.
   * Throws a `SyntaxError` if it is not JSON and a `ZodError` if it does not
   * match, e.g. when the model called tools instead of answering, or was cut
   * off.
   */
  decode(response: ChatResponse): z.output<S> {
    return this.schema.parse(JSON.parse(messageText(response.message)));
  }

  /** @internal The wire's `response_format`. */
  toWire(): ResponseFormatWire {
    const wire: ResponseFormatWire = {
      type: "json_schema",
      json_schema: { name: this.name, schema: toJsonSchema(this.schema) },
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
