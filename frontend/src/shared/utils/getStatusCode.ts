import type { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import type { SerializedError } from "@reduxjs/toolkit";

export function getStatusCode(
  error: FetchBaseQueryError | SerializedError | undefined
): number | undefined {
  if (!error) return undefined;
  if ("status" in error && typeof error.status === "number") return error.status;
  return undefined;
}
