import {createApi} from "@reduxjs/toolkit/query/react";

import {setCredentials} from "../model/authSlice";
import {baseQueryWithReauth} from "@/shared/api/baseApi.ts";
import {
  adminAuthPlatformStatisticsSchema,
  adminAuthUserStatisticsSchema,
} from "@/features/auth/model/schemas.ts";
import type {
  AdminAuthPlatformStatistics,
  AdminAuthUserStatistics,
} from "@/features/auth/model/types.ts";

const authApi = createApi({
  reducerPath: "authApi",
  baseQuery: baseQueryWithReauth,
  endpoints: (builder) => ({
    login: builder.mutation<
      { accessToken: string },
      { email: string; password: string }
    >({
      query: (body) => ({
        url: "/auth/login",
        method: "POST",
        body,
      }),
      async onQueryStarted(_, {dispatch, queryFulfilled}) {
        const {data} = await queryFulfilled;
        dispatch(setCredentials(data.accessToken));
      },
    }),

    register: builder.mutation<
      { userId: string; email: string },
      { email: string; password: string }
    >({
      query: (body) => ({
        url: "/auth/register",
        method: "POST",
        body,
      }),
    }),


    logout: builder.mutation<void, void>({
      query: () => ({
        url: "/auth/logout",
        method: "POST",
      }),
    }),

    verifyEmail: builder.mutation<{ status: string }, { token: string }>({
      query: ({token}) => ({
        url: "/auth/verify-email",
        method: "POST",
        body: {token},
      }),
    }),

    requestPasswordReset: builder.mutation<void, { email: string }>({
      query: (body) => ({
        url: "/auth/request-password-reset",
        method: "POST",
        body,
      }),
    }),

    resetPassword: builder.mutation<void, { token: string; newPassword: string }>({
      query: (body) => ({
        url: "/auth/reset-password",
        method: "POST",
        body,
      }),
    }),

    changePassword: builder.mutation<
      void,
      { oldEmail: string; oldPassword: string; newPassword: string }
    >({
      query: (body) => ({
        url: "/auth/change-password",
        method: "POST",
        body,
      }),
    }),

    getAdminPlatformStatistics: builder.query<AdminAuthPlatformStatistics, void>({
      query: () => "/auth/admin/statistics/platform",
      transformResponse: (response: unknown) => adminAuthPlatformStatisticsSchema.parse(response),
    }),

    getAdminUserStatistics: builder.query<AdminAuthUserStatistics, string>({
      query: (userId) => `/auth/admin/users/${userId}/statistics`,
      transformResponse: (response: unknown) => adminAuthUserStatisticsSchema.parse(response),
    }),
  }),
});

export default authApi;
