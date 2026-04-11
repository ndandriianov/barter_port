import { z } from "zod";

export const reviewContextTypeSchema = z.enum(["item-only", "offer-only", "offer+item"]);

export const reviewEligibilityReasonSchema = z.enum([
  "deal_not_completed",
  "forbidden_not_receiver",
  "receiver_missing",
  "provider_missing",
  "same_provider_and_receiver",
  "already_reviewed",
]);

export const offerRefSchema = z.object({
  offerId: z.string(),
});

export const dealItemRefSchema = z.object({
  dealId: z.string(),
  itemId: z.string(),
});

export const reviewSchema = z.object({
  id: z.string(),
  dealId: z.string(),
  authorId: z.string(),
  providerId: z.string(),
  rating: z.number().int().min(1).max(5),
  comment: z.string().optional(),
  createdAt: z.string(),
  updatedAt: z.string().optional(),
  offerRef: offerRefSchema.nullable().optional(),
  itemRef: dealItemRefSchema.nullable().optional(),
});

export const reviewsResponseSchema = z.array(reviewSchema);

export const reviewRatingBreakdownSchema = z.object({
  rating1: z.number().int().min(0),
  rating2: z.number().int().min(0),
  rating3: z.number().int().min(0),
  rating4: z.number().int().min(0),
  rating5: z.number().int().min(0),
});

export const reviewSummarySchema = z.object({
  count: z.number().int().min(0),
  avgRating: z.number().min(0).max(5),
  ratingBreakdown: reviewRatingBreakdownSchema,
});

export const reviewEligibilitySchema = z.object({
  canCreate: z.boolean(),
  contextType: reviewContextTypeSchema,
  providerId: z.string().nullable().optional(),
  offerRef: offerRefSchema.nullable().optional(),
  itemRef: dealItemRefSchema.nullable().optional(),
  reason: reviewEligibilityReasonSchema.nullable().optional(),
});

export const pendingDealReviewSchema = z.object({
  dealId: z.string(),
  contextType: reviewContextTypeSchema,
  providerId: z.string().nullable().optional(),
  offerRef: offerRefSchema.nullable().optional(),
  itemRef: dealItemRefSchema.nullable().optional(),
  canCreate: z.boolean(),
  reason: reviewEligibilityReasonSchema.nullable().optional(),
});

export const pendingDealReviewsResponseSchema = z.array(pendingDealReviewSchema);

export const createReviewRequestSchema = z.object({
  rating: z.number().int().min(1).max(5),
  comment: z.string().optional(),
});

export const updateReviewRequestSchema = z.object({
  rating: z.number().int().min(1).max(5).optional(),
  comment: z.string().optional(),
}).refine((value) => value.rating !== undefined || value.comment !== undefined, {
  message: "At least one field must be provided",
});
