import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "@/shared/api/baseApi.ts";
import { myStatisticsSchema } from "../model/schemas";
import type { MyStatistics } from "../model/types";

const statisticsApi = createApi({
  reducerPath: "statisticsApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Statistics"],
  endpoints: (builder) => ({
    getMyStatistics: builder.query<MyStatistics, void>({
      query: () => "/me/statistics",
      transformResponse: (response: unknown) => myStatisticsSchema.parse(response),
      providesTags: ["Statistics"],
    }),
  }),
});

export default statisticsApi;
