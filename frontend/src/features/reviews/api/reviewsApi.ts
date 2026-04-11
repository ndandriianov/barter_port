import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "@/shared/api/baseApi.ts";
import {
  createReviewRequestSchema,
  pendingDealReviewsResponseSchema,
  reviewEligibilitySchema,
  reviewsResponseSchema,
  reviewSchema,
  reviewSummarySchema,
  updateReviewRequestSchema,
} from "@/features/reviews/model/schemas.ts";
import type {
  CreateReviewRequest,
  PendingDealReviewsResponse,
  Review,
  ReviewEligibility,
  ReviewsResponse,
  ReviewSummary,
  UpdateReviewRequest,
} from "@/features/reviews/model/types.ts";

const reviewsApi = createApi({
  reducerPath: "reviewsApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Reviews", "PendingReviews", "ReviewSummary"],
  endpoints: (builder) => ({
    getOfferReviews: builder.query<ReviewsResponse, string>({
      query: (offerId) => `/offers/${offerId}/reviews`,
      transformResponse: (response: unknown) => reviewsResponseSchema.parse(response),
      providesTags: ["Reviews"],
    }),

    getOfferReviewsSummary: builder.query<ReviewSummary, string>({
      query: (offerId) => `/offers/${offerId}/reviews-summary`,
      transformResponse: (response: unknown) => reviewSummarySchema.parse(response),
      providesTags: ["ReviewSummary"],
    }),

    getProviderReviews: builder.query<ReviewsResponse, string>({
      query: (providerId) => `/providers/${providerId}/reviews`,
      transformResponse: (response: unknown) => reviewsResponseSchema.parse(response),
      providesTags: ["Reviews"],
    }),

    getProviderReviewsSummary: builder.query<ReviewSummary, string>({
      query: (providerId) => `/providers/${providerId}/reviews-summary`,
      transformResponse: (response: unknown) => reviewSummarySchema.parse(response),
      providesTags: ["ReviewSummary"],
    }),

    getAuthorReviews: builder.query<ReviewsResponse, string>({
      query: (authorId) => `/authors/${authorId}/reviews`,
      transformResponse: (response: unknown) => reviewsResponseSchema.parse(response),
      providesTags: ["Reviews"],
    }),

    getReviewById: builder.query<Review, string>({
      query: (reviewId) => `/reviews/${reviewId}`,
      transformResponse: (response: unknown) => reviewSchema.parse(response),
      providesTags: ["Reviews"],
    }),

    updateReview: builder.mutation<Review, { reviewId: string; body: UpdateReviewRequest }>({
      query: ({ reviewId, body }) => ({
        url: `/reviews/${reviewId}`,
        method: "PATCH",
        body: updateReviewRequestSchema.parse(body),
      }),
      transformResponse: (response: unknown) => reviewSchema.parse(response),
      invalidatesTags: ["Reviews", "PendingReviews", "ReviewSummary"],
    }),

    deleteReview: builder.mutation<void, string>({
      query: (reviewId) => ({
        url: `/reviews/${reviewId}`,
        method: "DELETE",
      }),
      invalidatesTags: ["Reviews", "PendingReviews", "ReviewSummary"],
    }),

    getDealReviews: builder.query<ReviewsResponse, string>({
      query: (dealId) => `/deals/${dealId}/reviews`,
      transformResponse: (response: unknown) => reviewsResponseSchema.parse(response),
      providesTags: ["Reviews"],
    }),

    getDealPendingReviews: builder.query<PendingDealReviewsResponse, string>({
      query: (dealId) => `/deals/${dealId}/reviews-pending`,
      transformResponse: (response: unknown) => pendingDealReviewsResponseSchema.parse(response),
      providesTags: ["PendingReviews"],
    }),

    getDealItemReviewEligibility: builder.query<ReviewEligibility, { dealId: string; itemId: string }>({
      query: ({ dealId, itemId }) => `/deals/${dealId}/items/${itemId}/reviews/eligibility`,
      transformResponse: (response: unknown) => reviewEligibilitySchema.parse(response),
      providesTags: ["PendingReviews"],
    }),

    getDealItemReviews: builder.query<ReviewsResponse, { dealId: string; itemId: string }>({
      query: ({ dealId, itemId }) => `/deals/${dealId}/items/${itemId}/reviews`,
      transformResponse: (response: unknown) => reviewsResponseSchema.parse(response),
      providesTags: ["Reviews"],
    }),

    createDealItemReview: builder.mutation<Review, { dealId: string; itemId: string; body: CreateReviewRequest }>({
      query: ({ dealId, itemId, body }) => ({
        url: `/deals/${dealId}/items/${itemId}/reviews`,
        method: "POST",
        body: createReviewRequestSchema.parse(body),
      }),
      transformResponse: (response: unknown) => reviewSchema.parse(response),
      invalidatesTags: ["Reviews", "PendingReviews", "ReviewSummary"],
    }),
  }),
});

export default reviewsApi;
