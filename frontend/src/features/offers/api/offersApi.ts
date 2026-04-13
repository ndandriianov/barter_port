import {createApi} from "@reduxjs/toolkit/query/react";
import {baseQueryWithReauth} from "@/shared/api/baseApi.ts";
import {getOffersResponseSchema, offerSchema} from "../model/schemas.ts";
import type {
  CreateOfferRequest,
  GetOffersParams,
  GetOffersResponse,
  Offer,
  UpdateOfferRequest,
} from "../model/types.ts";

const offersApi = createApi({
  reducerPath: "offersApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Offers"],
  endpoints: (builder) => ({
    getOffers: builder.query<GetOffersResponse, GetOffersParams>({
      query: (params) => ({
        url: "/offers",
        params,
      }),

      transformResponse: (response: unknown) => {
        console.log("Raw response from /offers:", response);
        return getOffersResponseSchema.parse(response);
      },

      providesTags: ["Offers"],
    }),

    getOfferById: builder.query<Offer, string>({
      query: (offerId) => `/offers/${offerId}`,
      transformResponse: (response: unknown) => offerSchema.parse(response),
      providesTags: (_result, _error, offerId) => [{type: "Offers", id: offerId}],
    }),

    createOffer: builder.mutation<void, CreateOfferRequest>({
      query: ({ photos = [], ...body }) => {
        if (photos.length === 0) {
          return {
            url: "/offers",
            method: "POST",
            body,
          };
        }

        const formData = new FormData();
        formData.append("name", body.name);
        formData.append("description", body.description);
        formData.append("action", body.action);
        formData.append("type", body.type);

        for (const photo of photos) {
          formData.append("photos", photo);
        }

        return {
          url: "/offers",
          method: "POST",
          body: formData,
        };
      },
      invalidatesTags: ["Offers"],
    }),

    updateOffer: builder.mutation<Offer, { offerId: string; body: UpdateOfferRequest }>({
      query: ({ offerId, body }) => {
        const hasPhotoChanges = (body.photos?.length ?? 0) > 0 || (body.deletePhotoIds?.length ?? 0) > 0;

        if (!hasPhotoChanges) {
          return {
            url: `/offers/${offerId}`,
            method: "PATCH",
            body,
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
      invalidatesTags: (_result, _error, { offerId }) => ["Offers", {type: "Offers", id: offerId}],
    }),

    deleteOffer: builder.mutation<void, string>({
      query: (offerId) => ({
        url: `/offers/${offerId}`,
        method: "DELETE",
      }),
      invalidatesTags: (_result, _error, offerId) => ["Offers", {type: "Offers", id: offerId}],
    }),
  }),
});

export default offersApi;
