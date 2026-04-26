import type { OfferGroup } from "@/features/offer-groups/model/types.ts";
import type { OfferAction } from "@/features/offers/model/types.ts";

export function getOfferGroupOwnerId(group: OfferGroup): string | undefined {
  return group.units[0]?.offers[0]?.authorId;
}

export function getOfferGroupOwnerName(group: OfferGroup): string {
  return group.units[0]?.offers[0]?.authorName?.trim() || "Имя не указано";
}

export function getOfferGroupVariantCount(group: OfferGroup): number {
  return group.units.reduce((total, unit) => total + unit.offers.length, 0);
}

export function formatOfferGroupDraftDealsCount(count: number): string {
  const mod10 = count % 10;
  const mod100 = count % 100;

  if (mod10 === 1 && mod100 !== 11) {
    return `${count} черновик`;
  }

  if (mod10 >= 2 && mod10 <= 4 && (mod100 < 12 || mod100 > 14)) {
    return `${count} черновика`;
  }

  return `${count} черновиков`;
}

export function getOfferGroupUnitActions(group: OfferGroup): OfferAction[] {
  return group.units
    .map((unit) => unit.offers[0]?.action)
    .filter((action): action is OfferAction => Boolean(action));
}

export function getOfferGroupUniformAction(group: OfferGroup): OfferAction | null {
  const actions = new Set(getOfferGroupUnitActions(group));
  if (actions.size !== 1) {
    return null;
  }

  return Array.from(actions)[0] ?? null;
}
