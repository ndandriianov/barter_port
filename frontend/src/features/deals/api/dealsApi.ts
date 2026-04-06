import {createApi} from "@reduxjs/toolkit/query/react";
import {baseQueryWithReauth} from "@/shared/api/baseApi.ts";
import {
  addDealItemRequestSchema,
  changeDealStatusRequestSchema,
  confirmDraftDealResponseSchema,
  createDraftDealResponseSchema,
  dealSchema,
  draftSchema,
  getDealJoinRequestsResponseSchema,
  getDealStatusVotesResponseSchema,
  getDealsResponseSchema,
  getMyDraftDealsResponseSchema,
  itemSchema,
} from "@/features/deals/model/schemas.ts";
import type {
  AddDealItemRequest,
  ChangeDealStatusRequest,
  ConfirmDraftDealResponse,
  CreateDraftDealRequest,
  CreateDraftDealResponse,
  Deal,
  GetDealJoinRequestsResponse,
  GetDealStatusVotesResponse,
  Draft,
  GetDealsParams,
  GetDealsResponse,
  GetMyDraftDealsParams,
  GetMyDraftDealsResponse,
  Item,
  UpdateDealItemRequest,
} from "@/features/deals/model/types.ts";

const dealsApi = createApi({
  reducerPath: "dealsApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Deals", "DraftDeals", "DealJoins", "DealStatusVotes"],
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
  }),
});

export default dealsApi;

