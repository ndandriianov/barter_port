import {z} from "zod";
import {
  addDealItemRequestSchema,
  adminDealsPlatformStatisticsSchema,
  adminDealsUserStatisticsSchema,
  changeDealStatusRequestSchema,
  confirmDraftDealResponseSchema,
  createDraftDealResponseSchema,
  dealSchema,
  dealStatusSchema,
  dealIdAndParticipantsSchema,
  dealJoinRequestSchema,
  draftSchema,
  failureMaterialsSchema,
  failureModerationDealsResponseSchema,
  failureResolutionSchema,
  failureVoteSchema,
  getDealJoinRequestsResponseSchema,
  getFailureVotesResponseSchema,
  getDealStatusVotesResponseItemSchema,
  getDealStatusVotesResponseSchema,
  getDealsResponseSchema,
  getMyDraftDealsResponseSchema,
  itemSchema,
  moderatorResolutionForFailureRequestSchema,
  offerWithInfoSchema,
  updateDealItemRequestSchema,
  userConfirmSchema,
  voteForFailureRequestSchema,
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
export type AddDealItemRequest = z.Infer<typeof addDealItemRequestSchema>;
export type UpdateDealItemRequest = z.Infer<typeof updateDealItemRequestSchema>;
export type DealIdAndParticipants = z.Infer<typeof dealIdAndParticipantsSchema>;
export type GetDealsResponse = z.Infer<typeof getDealsResponseSchema>;
export type DealJoinRequest = z.Infer<typeof dealJoinRequestSchema>;
export type GetDealJoinRequestsResponse = z.Infer<typeof getDealJoinRequestsResponseSchema>;
export type DealStatusVote = z.Infer<typeof getDealStatusVotesResponseItemSchema>;
export type GetDealStatusVotesResponse = z.Infer<typeof getDealStatusVotesResponseSchema>;
export type Item = z.Infer<typeof itemSchema>;
export type Deal = z.Infer<typeof dealSchema>;
export type VoteForFailureRequest = z.Infer<typeof voteForFailureRequestSchema>;
export type FailureVote = z.Infer<typeof failureVoteSchema>;
export type GetFailureVotesResponse = z.Infer<typeof getFailureVotesResponseSchema>;
export type ModeratorResolutionForFailureRequest = z.Infer<typeof moderatorResolutionForFailureRequestSchema>;
export type FailureResolution = z.Infer<typeof failureResolutionSchema>;
export type FailureMaterials = z.Infer<typeof failureMaterialsSchema>;
export type FailureModerationDealsResponse = z.Infer<typeof failureModerationDealsResponseSchema>;
export type AdminDealsPlatformStatistics = z.Infer<typeof adminDealsPlatformStatisticsSchema>;
export type AdminDealsUserStatistics = z.Infer<typeof adminDealsUserStatisticsSchema>;
