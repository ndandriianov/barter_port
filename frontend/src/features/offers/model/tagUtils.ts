export function normalizeOfferTags(tags: string[]): string[] {
  const seen = new Set<string>();
  const result: string[] = [];

  for (const tag of tags) {
    const normalized = tag.trim().toLowerCase();
    if (!normalized || seen.has(normalized)) {
      continue;
    }
    seen.add(normalized);
    result.push(normalized);
  }

  return result.sort((left, right) => left.localeCompare(right));
}

export function parseOfferTagsInput(value: string): string[] {
  return normalizeOfferTags(
    value
      .split(",")
      .map((item) => item.trim())
      .filter(Boolean),
  );
}
