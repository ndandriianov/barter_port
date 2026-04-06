import {z} from "zod";
import {
  changeDealStatusRequestSchema,
  confirmDraftDealResponseSchema,
  createDraftDealResponseSchema,
  dealSchema,
  dealStatusSchema,
  dealIdAndParticipantsSchema,
  dealJoinRequestSchema,
  draftSchema,
  getDealJoinRequestsResponseSchema,
  getDealStatusVotesResponseItemSchema,
  getDealStatusVotesResponseSchema,
  getDealsResponseSchema,
  getMyDraftDealsResponseSchema,
  itemSchema,
  offerWithInfoSchema,
  updateDealItemRequestSchema,
  userConfirmSchema,
  draftIdAndParticipantsSchema,
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
export type DraftIdAndParticipants = z.Infer<typeof draftIdAndParticipantsSchema>;
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

export type DealStatus = z.Infer<typeof dealStatusSchema>;
export type ChangeDealStatusRequest = z.Infer<typeof changeDealStatusRequestSchema>;
export type UpdateDealItemRequest = z.Infer<typeof updateDealItemRequestSchema>;
export type DealIdAndParticipants = z.Infer<typeof dealIdAndParticipantsSchema>;
export type GetDealsResponse = z.Infer<typeof getDealsResponseSchema>;
export type DealJoinRequest = z.Infer<typeof dealJoinRequestSchema>;
export type GetDealJoinRequestsResponse = z.Infer<typeof getDealJoinRequestsResponseSchema>;
export type DealStatusVote = z.Infer<typeof getDealStatusVotesResponseItemSchema>;
export type GetDealStatusVotesResponse = z.Infer<typeof getDealStatusVotesResponseSchema>;
export type Item = z.Infer<typeof itemSchema>;
export type Deal = z.Infer<typeof dealSchema>;

