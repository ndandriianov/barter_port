import {createApi} from "@reduxjs/toolkit/query/react";

import {setCredentials} from "../model/authSlice";
import {baseQueryWithReauth} from "@/shared/api/baseApi.ts";

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

    me: builder.query<{ userId: string }, void>({
      query: () => "/auth/me",
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
  }),
});

export default authApi;