export function getUserDisplayName(name: string | null | undefined, userId: string): string {
  const normalizedName = name?.trim();
  return normalizedName || userId.slice(0, 8);
}
