import type { Offer, OfferReportsForOffer } from "@/features/offers/model/types.ts";

export type OfferModerationState = "hidden" | "pending" | "reported" | null;

export function getOfferModerationState(
  offer: Pick<Offer, "isHidden" | "modificationBlocked">,
  reports?: Pick<OfferReportsForOffer, "reports"> | null,
): OfferModerationState {
  if (offer.isHidden) {
    return "hidden";
  }

  if (reports?.reports.some((thread) => thread.report.status === "Pending")) {
    return "pending";
  }

  if ((reports?.reports.length ?? 0) > 0 || offer.modificationBlocked) {
    return "reported";
  }

  return null;
}

export function getOfferModerationLabel(state: OfferModerationState): string | null {
  switch (state) {
    case "hidden":
      return "Скрыто";
    case "pending":
      return "На модерации";
    case "reported":
      return "Есть жалобы";
    default:
      return null;
  }
}

