import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "@/shared/api/baseApi.ts";
import {
  createOfferGroupDraftResponseSchema,
  createOfferGroupResponseSchema,
  listOfferGroupsResponseSchema,
  offerGroupSchema,
} from "@/features/offer-groups/model/schemas.ts";
import type {
  CreateOfferGroupDraftRequest,
  CreateOfferGroupDraftResponse,
  CreateOfferGroupRequest,
  CreateOfferGroupResponse,
  ListOfferGroupsResponse,
  OfferGroup,
} from "@/features/offer-groups/model/types.ts";

const offerGroupsApi = createApi({
  reducerPath: "offerGroupsApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["OfferGroups"],
  endpoints: (builder) => ({
    getOfferGroups: builder.query<ListOfferGroupsResponse, void>({
      query: () => "/offer-groups",
      transformResponse: (response: unknown) => listOfferGroupsResponseSchema.parse(response),
      providesTags: ["OfferGroups"],
    }),

    getOfferGroupById: builder.query<OfferGroup, string>({
      query: (offerGroupId) => `/offer-groups/${offerGroupId}`,
      transformResponse: (response: unknown) => offerGroupSchema.parse(response),
      providesTags: (_result, _error, offerGroupId) => [{ type: "OfferGroups", id: offerGroupId }],
    }),

    createOfferGroup: builder.mutation<CreateOfferGroupResponse, CreateOfferGroupRequest>({
      query: (body) => ({
        url: "/offer-groups",
        method: "POST",
        body,
      }),
      transformResponse: (response: unknown) => createOfferGroupResponseSchema.parse(response),
      invalidatesTags: ["OfferGroups"],
    }),

    createDraftFromOfferGroup: builder.mutation<
      CreateOfferGroupDraftResponse,
      { offerGroupId: string; body: CreateOfferGroupDraftRequest }
    >({
      query: ({ offerGroupId, body }) => ({
        url: `/offer-groups/${offerGroupId}/drafts`,
        method: "POST",
        body,
      }),
      transformResponse: (response: unknown) => createOfferGroupDraftResponseSchema.parse(response),
      invalidatesTags: ["OfferGroups"],
    }),
  }),
});

export default offerGroupsApi;
