import {createApi} from "@reduxjs/toolkit/query/react";
import {baseQueryWithReauth} from "@/shared/api/baseApi.ts";
import {
  addDealItemRequestSchema,
  changeDealStatusRequestSchema,
  confirmDraftDealResponseSchema,
  createDraftDealResponseSchema,
  dealSchema,
  draftSchema,
  failureMaterialsSchema,
  failureModerationDealsResponseSchema,
  failureResolutionSchema,
  getDealJoinRequestsResponseSchema,
  getFailureVotesResponseSchema,
  getDealStatusVotesResponseSchema,
  getDealsResponseSchema,
  getMyDraftDealsResponseSchema,
  itemSchema,
  moderatorResolutionForFailureRequestSchema,
  voteForFailureRequestSchema,
} from "@/features/deals/model/schemas.ts";
import type {
  AddDealItemRequest,
  ChangeDealStatusRequest,
  ConfirmDraftDealResponse,
  CreateDraftDealRequest,
  CreateDraftDealResponse,
  Deal,
  FailureMaterials,
  FailureModerationDealsResponse,
  FailureResolution,
  GetFailureVotesResponse,
  GetDealJoinRequestsResponse,
  GetDealStatusVotesResponse,
  Draft,
  GetDealsParams,
  GetDealsResponse,
  GetMyDraftDealsParams,
  GetMyDraftDealsResponse,
  Item,
  ModeratorResolutionForFailureRequest,
  UpdateDealItemRequest,
  VoteForFailureRequest,
} from "@/features/deals/model/types.ts";

const dealsApi = createApi({
  reducerPath: "dealsApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Deals", "DraftDeals", "DealJoins", "DealStatusVotes", "FailureReview", "FailureVotes", "FailureResolution", "FailureMaterials"],
  endpoints: (builder) => ({
    createDraftDeal: builder.mutation<CreateDraftDealResponse, CreateDraftDealRequest>({
      query: (body) => ({
        url: "/deals/drafts",
        method: "POST",
        body,
      }),
      transformResponse: (response: unknown) => createDraftDealResponseSchema.parse(response),
      invalidatesTags: ["DraftDeals"],
    }),

    getMyDraftDeals: builder.query<GetMyDraftDealsResponse, GetMyDraftDealsParams | void>({
      query: (params) =>
        params
          ? {
              url: "/deals/drafts",
              params,
            }
          : "/deals/drafts",
      transformResponse: (response: unknown) => getMyDraftDealsResponseSchema.parse(response),
      providesTags: ["DraftDeals"],
    }),

    getDraftDealById: builder.query<Draft, string>({
      query: (draftId) => `/deals/drafts/${draftId}`,
      transformResponse: (response: unknown) => draftSchema.parse(response),
      providesTags: (_result, _error, draftId) => [{type: "DraftDeals", id: draftId}],
    }),

    confirmDraftDeal: builder.mutation<ConfirmDraftDealResponse, string>({
      query: (draftId) => ({
        url: `/deals/drafts/${draftId}`,
        method: "PATCH",
      }),
      transformResponse: (response: unknown) => confirmDraftDealResponseSchema.parse(response),
      invalidatesTags: (_result, _error, draftId) => [
        {type: "DraftDeals", id: draftId},
        "Deals",
      ],
    }),

    cancelDraftDeal: builder.mutation<void, string>({
      query: (draftId) => ({
        url: `/deals/drafts/${draftId}/cancel`,
        method: "PATCH",
      }),
      invalidatesTags: (_result, _error, draftId) => [
        {type: "DraftDeals", id: draftId},
        "DraftDeals",
      ],
    }),

    deleteDraftDeal: builder.mutation<void, string>({
      query: (draftId) => ({
        url: `/deals/drafts/${draftId}`,
        method: "DELETE",
      }),
      invalidatesTags: (_result, _error, draftId) => [
        {type: "DraftDeals", id: draftId},
        "DraftDeals",
      ],
    }),

    getDeals: builder.query<GetDealsResponse, GetDealsParams | void>({
      query: (params) =>
        params
          ? {
              url: "/deals",
              params,
            }
          : "/deals",
      transformResponse: (response: unknown) => getDealsResponseSchema.parse(response),
      providesTags: ["Deals"],
    }),

    getDealById: builder.query<Deal, string>({
      query: (dealId) => `/deals/${dealId}`,
      transformResponse: (response: unknown) => dealSchema.parse(response),
      providesTags: (_result, _error, dealId) => [{type: "Deals", id: dealId}],
    }),

    updateDeal: builder.mutation<Deal, { dealId: string; name: string }>({
      query: ({ dealId, name }) => ({
        url: `/deals/${dealId}`,
        method: "PATCH",
        body: { name },
      }),
      transformResponse: (response: unknown) => dealSchema.parse(response),
      invalidatesTags: (_result, _error, { dealId }) => [
        {type: "Deals", id: dealId},
        "Deals",
      ],
    }),

    joinDeal: builder.mutation<void, string>({
      query: (dealId) => ({
        url: `/deals/${dealId}/joins`,
        method: "POST",
      }),
      invalidatesTags: (_result, _error, dealId) => [
        {type: "Deals", id: dealId},
        {type: "DealJoins", id: dealId},
        "Deals",
      ],
    }),

    getDealJoins: builder.query<GetDealJoinRequestsResponse, string>({
      query: (dealId) => `/deals/${dealId}/joins`,
      transformResponse: (response: unknown) => getDealJoinRequestsResponseSchema.parse(response),
      providesTags: (_result, _error, dealId) => [{type: "DealJoins", id: dealId}],
    }),

    getDealStatusVotes: builder.query<GetDealStatusVotesResponse, string>({
      query: (dealId) => `/deals/${dealId}/status`,
      transformResponse: (response: unknown) => getDealStatusVotesResponseSchema.parse(response),
      providesTags: (_result, _error, dealId) => [{type: "DealStatusVotes", id: dealId}],
    }),

    leaveDeal: builder.mutation<void, string>({
      query: (dealId) => ({
        url: `/deals/${dealId}/joins`,
        method: "DELETE",
      }),
      invalidatesTags: (_result, _error, dealId) => [
        {type: "Deals", id: dealId},
        {type: "DealJoins", id: dealId},
        "Deals",
      ],
    }),

    processJoinRequest: builder.mutation<void, { dealId: string; userId: string; accept: boolean }>({
      query: ({ dealId, userId, accept }) => ({
        url: `/deals/${dealId}/joins/${userId}`,
        method: "POST",
        params: {accept},
      }),
      invalidatesTags: (_result, _error, { dealId }) => [
        {type: "Deals", id: dealId},
        {type: "DealJoins", id: dealId},
        "Deals",
      ],
    }),

    changeDealStatus: builder.mutation<Deal, { dealId: string; body: ChangeDealStatusRequest }>({
      query: ({ dealId, body }) => ({
        url: `/deals/${dealId}/status`,
        method: "PATCH",
        body: changeDealStatusRequestSchema.parse(body),
      }),
      transformResponse: (response: unknown) => dealSchema.parse(response),
      invalidatesTags: (_result, _error, { dealId }) => [
        {type: "Deals", id: dealId},
        {type: "DealStatusVotes", id: dealId},
        "Deals",
      ],
    }),

    addDealItem: builder.mutation<Deal, { dealId: string; body: AddDealItemRequest }>({
      query: ({ dealId, body }) => ({
        url: `/deals/${dealId}/items`,
        method: "POST",
        body: addDealItemRequestSchema.parse(body),
      }),
      transformResponse: (response: unknown) => dealSchema.parse(response),
      invalidatesTags: (_result, _error, { dealId }) => [
        {type: "Deals", id: dealId},
        "Deals",
      ],
    }),

    updateDealItem: builder.mutation<Item, { dealId: string; itemId: string; body: UpdateDealItemRequest }>({
      query: ({ dealId, itemId, body }) => ({
        url: `/deals/${dealId}/items/${itemId}`,
        method: "PATCH",
        body,
      }),
      transformResponse: (response: unknown) => itemSchema.parse(response),
      invalidatesTags: (_result, _error, { dealId }) => [{type: "Deals", id: dealId}],
    }),

    getDealsForFailureReview: builder.query<FailureModerationDealsResponse, void>({
      query: () => "/deals/failures/review",
      transformResponse: (response: unknown) => failureModerationDealsResponseSchema.parse(response),
      providesTags: ["FailureReview"],
    }),

    voteForFailure: builder.mutation<void, { dealId: string; body: VoteForFailureRequest }>({
      query: ({ dealId, body }) => ({
        url: `/deals/failures/${dealId}/votes`,
        method: "POST",
        body: voteForFailureRequestSchema.parse(body),
      }),
      invalidatesTags: (_result, _error, { dealId }) => [
        { type: "FailureVotes", id: dealId },
        { type: "FailureResolution", id: dealId },
        { type: "FailureMaterials", id: dealId },
        "FailureReview",
      ],
    }),

    revokeVoteForFailure: builder.mutation<void, string>({
      query: (dealId) => ({
        url: `/deals/failures/${dealId}/votes`,
        method: "DELETE",
      }),
      invalidatesTags: (_result, _error, dealId) => [
        { type: "FailureVotes", id: dealId },
        { type: "FailureResolution", id: dealId },
        { type: "FailureMaterials", id: dealId },
        "FailureReview",
      ],
    }),

    getFailureVotes: builder.query<GetFailureVotesResponse, string>({
      query: (dealId) => `/deals/failures/${dealId}/votes`,
      transformResponse: (response: unknown) => getFailureVotesResponseSchema.parse(response),
      providesTags: (_result, _error, dealId) => [{ type: "FailureVotes", id: dealId }],
    }),

    getFailureMaterials: builder.query<FailureMaterials, string>({
      query: (dealId) => `/deals/failures/${dealId}/materials`,
      transformResponse: (response: unknown) => failureMaterialsSchema.parse(response),
      providesTags: (_result, _error, dealId) => [{ type: "FailureMaterials", id: dealId }],
    }),

    moderatorResolutionForFailure: builder.mutation<FailureResolution, { dealId: string; body: ModeratorResolutionForFailureRequest }>({
      query: ({ dealId, body }) => ({
        url: `/deals/failures/${dealId}/moderator-resolution`,
        method: "POST",
        body: moderatorResolutionForFailureRequestSchema.parse(body),
      }),
      transformResponse: (response: unknown) => failureResolutionSchema.parse(response),
      invalidatesTags: (_result, _error, { dealId }) => [
        { type: "FailureReview", id: dealId },
        { type: "FailureVotes", id: dealId },
        { type: "FailureResolution", id: dealId },
        { type: "FailureMaterials", id: dealId },
        { type: "Deals", id: dealId },
        "Deals",
        "FailureReview",
      ],
    }),

    getModeratorResolutionForFailure: builder.query<FailureResolution, string>({
      query: (dealId) => `/deals/failures/${dealId}/moderator-resolution`,
      transformResponse: (response: unknown) => failureResolutionSchema.parse(response),
      providesTags: (_result, _error, dealId) => [{ type: "FailureResolution", id: dealId }],
    }),
  }),
});

export default dealsApi;
