import {z} from "zod";
import {
  getOffersResponseSchema,
  offerActionSchema,
  offerSchema,
  offerTypeSchema,
  universalCursorSchema,
} from "@/features/offers/model/schemas.ts";

export type OfferAction = z.Infer<typeof offerActionSchema>;
export type OfferType = z.Infer<typeof offerTypeSchema>;

export type Offer = z.Infer<typeof offerSchema>;
export type UniversalCursor = z.Infer<typeof universalCursorSchema>;

export type SortType = "ByTime" | "ByPopularity";

export interface GetOffersParams {
  sort: SortType;
  cursor_created_at?: string;
  cursor_views?: number;
  cursor_id?: string;
  cursor_limit?: number;
}

export type GetOffersResponse = z.Infer<typeof getOffersResponseSchema>;

export interface CreateOfferRequest {
  name: string;
  description: string;
  action: OfferAction;
  type: OfferType;
}

