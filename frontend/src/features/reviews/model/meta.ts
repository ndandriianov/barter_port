import type { ReviewContextType, ReviewEligibilityReason } from "@/features/reviews/model/types.ts";

export const reviewContextLabels: Record<ReviewContextType, string> = {
  "item-only": "Отзыв о позиции",
  "offer-only": "Отзыв об offer",
  "offer+item": "Отзыв об offer и позиции",
};

export const reviewReasonLabels: Record<ReviewEligibilityReason, string> = {
  deal_not_completed: "Оставить отзыв можно только после завершения сделки",
  forbidden_not_receiver: "Отзыв может оставить только получатель этой позиции",
  receiver_missing: "Для позиции не указан получатель",
  provider_missing: "Для позиции не указан поставщик",
  same_provider_and_receiver: "Нельзя оставить отзыв самому себе",
  already_reviewed: "Отзыв по этому контексту уже оставлен",
};

export const formatReviewDate = (value: string) =>
  new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
