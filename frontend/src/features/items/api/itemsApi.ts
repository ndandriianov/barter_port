import {createApi} from "@reduxjs/toolkit/query/react";

import {baseQueryWithReauth} from "@/shared/api/baseApi.ts";
import type {
  GetItemsParams,
  GetItemsResponse,
  CreateItemRequest,
} from "../model/types.ts";

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
