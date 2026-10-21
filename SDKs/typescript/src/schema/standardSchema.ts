// The subset of Standard Schema V1 and Standard JSON Schema V1
// (https://standardschema.dev) this SDK reads, copied as the spec recommends so
// the SDK keeps no runtime dependency. Zod 4.2+ and ArkType implement both on
// their schemas; Valibot through `toStandardJsonSchema`.

/** A schema that validates values and converts to JSON Schema. */
export interface StandardSchemaWithJson<Output = unknown> {
  readonly "~standard": {
    readonly version: 1;
    readonly vendor: string;
    readonly validate: (value: unknown) => StandardResult<Output> | Promise<StandardResult<Output>>;
    readonly jsonSchema: {
      readonly input: (options: { readonly target: string }) => Record<string, unknown>;
    };
    readonly types?: { readonly input: unknown; readonly output: Output } | undefined;
  };
}

export type StandardResult<Output> =
  | { readonly value: Output; readonly issues?: undefined }
  | { readonly issues: readonly StandardIssue[] };

export interface StandardIssue {
  readonly message: string;
  readonly path?: readonly (PropertyKey | { readonly key: PropertyKey })[] | undefined;
}
