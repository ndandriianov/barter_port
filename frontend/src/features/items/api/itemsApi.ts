import {createApi} from "@reduxjs/toolkit/query/react";
import {baseQueryWithReauth} from "@/shared/api/baseApi.ts";
import {getItemsResponseSchema} from "../model/schemas.ts";
import type {CreateItemRequest, GetItemsParams, GetItemsResponse} from "../model/types.ts";

const itemsApi = createApi({
  reducerPath: "itemsApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Items"],
  endpoints: (builder) => ({
    getItems: builder.query<GetItemsResponse, GetItemsParams>({
      query: (params) => ({
        url: "/items",
        params,
      }),

      transformResponse: (response: unknown) => {
        console.log("Raw response from /items:", response);
        return getItemsResponseSchema.parse(response);
      },

      providesTags: ["Items"],
    }),

    createItem: builder.mutation<void, CreateItemRequest>({
      query: (body) => ({
        url: "/items",
        method: "POST",
        body,
      }),
      invalidatesTags: ["Items"],
    }),
  }),
});

export default itemsApi;
