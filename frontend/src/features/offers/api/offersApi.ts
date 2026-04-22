import {createApi} from "@reduxjs/toolkit/query/react";
import {baseQueryWithReauth} from "@/shared/api/baseApi.ts";
import {
  getFavoriteOffersResponseSchema,
  getOffersResponseSchema,
  listTagsResponseSchema,
  listOfferReportsResponseSchema,
  offerReportDetailsSchema,
  offerReportSchema,
  offerReportsForOfferSchema,
  offerSchema,
} from "../model/schemas.ts";
import type {
  CreateOfferReportRequest,
  CreateOfferRequest,
  GetFavoriteOffersParams,
  GetFavoriteOffersResponse,
  GetOffersParams,
  GetOffersResponse,
  GetSubscribedOffersParams,
  ListOfferReportsResponse,
  ListTagsResponse,
  Offer,
  OfferReport,
  OfferReportDetails,
  OfferReportStatus,
  OfferReportsForOffer,
  ResolveOfferReportRequest,
  UpdateOfferRequest,
} from "../model/types.ts";

const offersApi = createApi({
  reducerPath: "offersApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Offers", "FavoriteOffers", "OfferReports", "AdminOfferReports", "Tags"],
  endpoints: (builder) => ({
    getOffers: builder.query<GetOffersResponse, GetOffersParams>({
      query: (params) => ({
        url: "/offers",
        params,
      }),
      transformResponse: (response: unknown) => getOffersResponseSchema.parse(response),
      providesTags: ["Offers"],
    }),

    getSubscribedOffers: builder.query<GetOffersResponse, GetSubscribedOffersParams>({
      query: (params) => ({
        url: "/offers/subscriptions",
        params,
      }),
      transformResponse: (response: unknown) => getOffersResponseSchema.parse(response),
      providesTags: ["Offers"],
    }),

    getFavoriteOffers: builder.query<GetFavoriteOffersResponse, GetFavoriteOffersParams | void>({
      query: (params) =>
        params
          ? {
              url: "/offers/favorites",
              params,
            }
          : "/offers/favorites",
      transformResponse: (response: unknown) => getFavoriteOffersResponseSchema.parse(response),
      providesTags: ["FavoriteOffers", "Offers"],
    }),

    getOfferById: builder.query<Offer, string>({
      query: (offerId) => `/offers/${offerId}`,
      transformResponse: (response: unknown) => offerSchema.parse(response),
      providesTags: (_result, _error, offerId) => [{type: "Offers", id: offerId}],
    }),

    viewOfferById: builder.mutation<void, string>({
      query: (offerId) => ({
        url: `/offers/${offerId}/view`,
        method: "POST",
      }),
      invalidatesTags: (_result, _error, offerId) => ["Offers", {type: "Offers", id: offerId}],
    }),

    addOfferToFavorites: builder.mutation<void, string>({
      query: (offerId) => ({
        url: `/offers/${offerId}/favorite`,
        method: "PUT",
      }),
      invalidatesTags: (_result, _error, offerId) => [
        "Offers",
        "FavoriteOffers",
        { type: "Offers", id: offerId },
      ],
    }),

    removeOfferFromFavorites: builder.mutation<void, string>({
      query: (offerId) => ({
        url: `/offers/${offerId}/favorite`,
        method: "DELETE",
      }),
      invalidatesTags: (_result, _error, offerId) => [
        "Offers",
        "FavoriteOffers",
        { type: "Offers", id: offerId },
      ],
    }),

    createOffer: builder.mutation<void, CreateOfferRequest>({
      query: ({ photos = [], tags = [], ...body }) => {
        if (photos.length === 0) {
          return {
            url: "/offers",
            method: "POST",
            body: {
              ...body,
              tags: tags.length > 0 ? tags : undefined,
            },
          };
        }

        const formData = new FormData();
        formData.append("name", body.name);
        formData.append("description", body.description);
        formData.append("action", body.action);
        formData.append("type", body.type);
        for (const tag of tags) {
          formData.append("tags", tag);
        }

        for (const photo of photos) {
          formData.append("photos", photo);
        }

        return {
          url: "/offers",
          method: "POST",
          body: formData,
        };
      },
      invalidatesTags: ["Offers", "Tags"],
    }),

    updateOffer: builder.mutation<Offer, { offerId: string; body: UpdateOfferRequest }>({
      query: ({ offerId, body }) => {
        const hasPhotoChanges = (body.photos?.length ?? 0) > 0 || (body.deletePhotoIds?.length ?? 0) > 0;

        if (!hasPhotoChanges) {
          return {
            url: `/offers/${offerId}`,
            method: "PATCH",
            body: {
              ...body,
              tags: body.tags,
            },
          };
        }

        const formData = new FormData();

        if (body.name !== undefined) {
          formData.append("name", body.name);
        }
        if (body.description !== undefined) {
          formData.append("description", body.description);
        }
        if (body.action !== undefined) {
          formData.append("action", body.action);
        }
        if (body.type !== undefined) {
          formData.append("type", body.type);
        }
        if (body.tags !== undefined) {
          for (const tag of body.tags) {
            formData.append("tags", tag);
          }
        }
        for (const photoId of body.deletePhotoIds ?? []) {
          formData.append("deletePhotoIds", photoId);
        }
        for (const photo of body.photos ?? []) {
          formData.append("photos", photo);
        }

        return {
          url: `/offers/${offerId}`,
          method: "PATCH",
          body: formData,
        };
      },
      transformResponse: (response: unknown) => offerSchema.parse(response),
      invalidatesTags: (_result, _error, { offerId }) => ["Offers", {type: "Offers", id: offerId}, "Tags"],
    }),

    deleteOffer: builder.mutation<void, string>({
      query: (offerId) => ({
        url: `/offers/${offerId}`,
        method: "DELETE",
      }),
      invalidatesTags: (_result, _error, offerId) => ["Offers", {type: "Offers", id: offerId}, "Tags"],
    }),

    listTags: builder.query<ListTagsResponse, void>({
      query: () => "/tags",
      transformResponse: (response: unknown) => listTagsResponseSchema.parse(response),
      providesTags: ["Tags"],
    }),

    deleteAdminTag: builder.mutation<void, string>({
      query: (name) => ({
        url: "/admin/tags",
        method: "DELETE",
        params: { name },
      }),
      invalidatesTags: ["Tags", "Offers"],
    }),

    createOfferReport: builder.mutation<OfferReport, { offerId: string; body: CreateOfferReportRequest }>({
      query: ({ offerId, body }) => ({
        url: `/offers/${offerId}/reports`,
        method: "POST",
        body,
      }),
      transformResponse: (response: unknown) => offerReportSchema.parse(response),
      invalidatesTags: (_result, _error, { offerId }) => [
        { type: "OfferReports", id: offerId },
        "OfferReports",
        "AdminOfferReports",
        "Offers",
        { type: "Offers", id: offerId },
      ],
    }),

    getOfferReports: builder.query<OfferReportsForOffer, string>({
      query: (offerId) => `/offers/${offerId}/reports`,
      transformResponse: (response: unknown) => offerReportsForOfferSchema.parse(response),
      providesTags: (_result, _error, offerId) => [{ type: "OfferReports", id: offerId }],
    }),

    listAdminOfferReports: builder.query<ListOfferReportsResponse, OfferReportStatus | void>({
      query: (status) =>
        status
          ? {
              url: "/admin/offer-reports",
              params: { status },
            }
          : "/admin/offer-reports",
      transformResponse: (response: unknown) => listOfferReportsResponseSchema.parse(response),
      providesTags: ["AdminOfferReports"],
    }),

    getAdminOfferReportById: builder.query<OfferReportDetails, string>({
      query: (reportId) => `/admin/offer-reports/${reportId}`,
      transformResponse: (response: unknown) => offerReportDetailsSchema.parse(response),
      providesTags: (_result, _error, reportId) => [{ type: "AdminOfferReports", id: reportId }],
    }),

    resolveAdminOfferReport: builder.mutation<OfferReport, { reportId: string; body: ResolveOfferReportRequest }>({
      query: ({ reportId, body }) => ({
        url: `/admin/offer-reports/${reportId}/resolution`,
        method: "POST",
        body,
      }),
      transformResponse: (response: unknown) => offerReportSchema.parse(response),
      invalidatesTags: (_result, _error, { reportId }) => [
        "AdminOfferReports",
        { type: "AdminOfferReports", id: reportId },
        "OfferReports",
        "Offers",
      ],
    }),
  }),
});

export default offersApi;
