/**
 * A tool the model may call, its arguments described by the JSON Schema
 * `parameters`, e.g. `toJsonSchema(ArgsSchema)`.
 */
export interface ToolDefinition {
  name: string;
  description?: string;
  parameters?: Record<string, unknown>;
}
