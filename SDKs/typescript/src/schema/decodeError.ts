import type { StandardIssue } from "./standardSchema";

/**
 * Thrown when a value does not decode with its schema: `issues` says why. A
 * text that is not JSON has no issues; its parse error is the `cause`.
 */
export class DecodeError extends Error {
  override readonly name = "DecodeError";
  readonly issues: readonly StandardIssue[];

  constructor(issues: readonly StandardIssue[], options?: ErrorOptions) {
    super(issues.map((issue) => issue.message).join("; ") || "not JSON", options);
    this.issues = issues;
  }
}
