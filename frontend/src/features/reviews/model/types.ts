import { z } from "zod";
import {
  createReviewRequestSchema,
  dealItemRefSchema,
  offerRefSchema,
  pendingDealReviewSchema,
  pendingDealReviewsResponseSchema,
  reviewContextTypeSchema,
  reviewEligibilityReasonSchema,
  reviewEligibilitySchema,
  reviewSchema,
  reviewsResponseSchema,
  reviewSummarySchema,
  updateReviewRequestSchema,
} from "@/features/reviews/model/schemas.ts";

export type ReviewContextType = z.Infer<typeof reviewContextTypeSchema>;
export type ReviewEligibilityReason = z.Infer<typeof reviewEligibilityReasonSchema>;
export type OfferRef = z.Infer<typeof offerRefSchema>;
export type DealItemRef = z.Infer<typeof dealItemRefSchema>;
export type Review = z.Infer<typeof reviewSchema>;
export type ReviewsResponse = z.Infer<typeof reviewsResponseSchema>;
export type ReviewSummary = z.Infer<typeof reviewSummarySchema>;
export type ReviewEligibility = z.Infer<typeof reviewEligibilitySchema>;
export type PendingDealReview = z.Infer<typeof pendingDealReviewSchema>;
export type PendingDealReviewsResponse = z.Infer<typeof pendingDealReviewsResponseSchema>;
export type CreateReviewRequest = z.Infer<typeof createReviewRequestSchema>;
export type UpdateReviewRequest = z.Infer<typeof updateReviewRequestSchema>;
