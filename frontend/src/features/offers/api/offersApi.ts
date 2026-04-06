import {createApi} from "@reduxjs/toolkit/query/react";
import {baseQueryWithReauth} from "@/shared/api/baseApi.ts";
import {getOffersResponseSchema} from "../model/schemas.ts";
import type {CreateOfferRequest, GetOffersParams, GetOffersResponse} from "../model/types.ts";

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

    createOffer: builder.mutation<void, CreateOfferRequest>({
      query: (body) => ({
        url: "/offers",
        method: "POST",
        body,
      }),
      invalidatesTags: ["Offers"],
    }),
  }),
});

export default offersApi;

