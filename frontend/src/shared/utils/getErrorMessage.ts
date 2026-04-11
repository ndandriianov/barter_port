type ErrorResponse = {
  message?: string | null;
};

export function getErrorMessage(error: unknown): string | undefined {
  if (!error) return undefined;

  if (typeof error === "object" && error !== null && "status" in error) {
    const data = (error as { data?: unknown }).data;
    if (data && typeof data === "object" && "message" in data) {
      const message = (data as ErrorResponse).message;
      if (typeof message === "string" && message.trim()) {
        return message;
      }
    }
  }

  if (typeof error === "object" && error !== null && "message" in error) {
    const message = (error as { message?: unknown }).message;
    if (typeof message === "string" && message.trim()) {
      return message;
    }
  }

  return undefined;
}
