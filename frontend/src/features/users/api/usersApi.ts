import {createApi} from "@reduxjs/toolkit/query/react";
import {baseQueryWithReauth} from "@/shared/api/baseApi.ts";
import {meSchema, userSchema} from "@/features/users/model/schemas.ts";
import type {Me, UpdateCurrentUserRequest, User} from "@/features/users/model/types.ts";

const usersApi = createApi({
  reducerPath: "usersApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["CurrentUser", "Users"],
  endpoints: (builder) => ({
    getCurrentUser: builder.query<Me, void>({
      query: () => "/users/me",
      transformResponse: (response: unknown) => meSchema.parse(response),
      providesTags: ["CurrentUser"],
    }),

    updateCurrentUser: builder.mutation<Me, UpdateCurrentUserRequest>({
      query: (body) => ({
        url: "/users/me",
        method: "PATCH",
        body,
      }),
      transformResponse: (response: unknown) => meSchema.parse(response),
      invalidatesTags: ["CurrentUser"],
    }),

    getUserById: builder.query<User, string>({
      query: (id) => `/users/${id}`,
      transformResponse: (response: unknown) => userSchema.parse(response),
      providesTags: (_result, _error, id) => [{type: "Users", id}],
    }),
  }),
});

export default usersApi;

