import type { OfferGroup } from "@/features/offer-groups/model/types.ts";

export function getOfferGroupOwnerId(group: OfferGroup): string | undefined {
  return group.units[0]?.offers[0]?.authorId;
}

export function getOfferGroupOwnerName(group: OfferGroup): string {
  return group.units[0]?.offers[0]?.authorName?.trim() || "Имя не указано";
}

export function getOfferGroupVariantCount(group: OfferGroup): number {
  return group.units.reduce((total, unit) => total + unit.offers.length, 0);
}
