import {z} from "zod";
import {
  confirmDraftDealResponseSchema,
  createDraftDealResponseSchema,
  dealSchema,
  draftSchema,
  getDealsResponseSchema,
  getMyDraftDealsResponseSchema,
  itemSchema,
  offerWithInfoSchema,
  userConfirmSchema,
} from "@/features/deals/model/schemas.ts";

export type OfferWithInfo = z.Infer<typeof offerWithInfoSchema>;
export type Draft = z.Infer<typeof draftSchema>;

export interface OfferIDAndQuantity {
  offerID: string;
  quantity: number;
}

export interface CreateDraftDealRequest {
  name?: string;
  description?: string;
  offers: OfferIDAndQuantity[];
}

export type CreateDraftDealResponse = z.Infer<typeof createDraftDealResponseSchema>;
export type GetMyDraftDealsResponse = z.Infer<typeof getMyDraftDealsResponseSchema>;

export interface GetMyDraftDealsParams {
  createdByMe?: boolean;
  participating?: boolean;
}

export type UserConfirm = z.Infer<typeof userConfirmSchema>;
export type ConfirmDraftDealResponse = z.Infer<typeof confirmDraftDealResponseSchema>;

export interface GetDealsParams {
  my?: boolean;
  open?: boolean;
}

export type GetDealsResponse = z.Infer<typeof getDealsResponseSchema>;
export type Item = z.Infer<typeof itemSchema>;
export type Deal = z.Infer<typeof dealSchema>;

