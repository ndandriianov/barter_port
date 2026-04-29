import {createApi} from "@reduxjs/toolkit/query/react";
import {baseQueryWithReauth} from "@/shared/api/baseApi.ts";
import {
  adminUsersPlatformStatisticsSchema,
  adminUsersUserStatisticsSchema,
  meSchema,
  reputationEventsResponseSchema,
  subscriptionsResponseSchema,
  userAvatarUploadSchema,
  userSchema,
} from "@/features/users/model/schemas.ts";
import type {
  AdminUsersPlatformStatistics,
  AdminUsersUserStatistics,
  Me,
  ReputationEvent,
  SubscribeRequest,
  SubscriptionsResponse,
  UpdateCurrentUserRequest,
  UploadCurrentUserAvatarResponse,
  User
} from "@/features/users/model/types.ts";

const usersApi = createApi({
  reducerPath: "usersApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["CurrentUser", "Users", "Subscriptions"],
  endpoints: (builder) => ({
    getCurrentUser: builder.query<Me, void>({
      query: () => "/users/me",
      transformResponse: (response: unknown) => meSchema.parse(response),
      providesTags: ["CurrentUser"],
    }),

    getCurrentUserReputationEvents: builder.query<ReputationEvent[], void>({
      query: () => "/users/reputation-events",
      transformResponse: (response: unknown) => reputationEventsResponseSchema.parse(response),
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

    uploadCurrentUserAvatar: builder.mutation<UploadCurrentUserAvatarResponse, FormData>({
      query: (body) => ({
        url: "/users/me/avatar",
        method: "POST",
        body,
      }),
      transformResponse: (response: unknown) => userAvatarUploadSchema.parse(response),
    }),

    getUserById: builder.query<User, string>({
      query: (id) => `/users/${id}`,
      transformResponse: (response: unknown) => userSchema.parse(response),
      providesTags: (_result, _error, id) => [{type: "Users", id}],
    }),

    getSubscriptions: builder.query<SubscriptionsResponse, void>({
      query: () => "/users/subscriptions",
      transformResponse: (response: unknown) => subscriptionsResponseSchema.parse(response),
      providesTags: ["Subscriptions"],
    }),

    getSubscriptionsByUserId: builder.query<SubscriptionsResponse, string>({
      query: (id) => `/users/subscriptions/${id}`,
      transformResponse: (response: unknown) => subscriptionsResponseSchema.parse(response),
      providesTags: (_result, _error, id) => [{type: "Subscriptions", id: `subscriptions:${id}`}],
    }),

    getSubscribers: builder.query<SubscriptionsResponse, void>({
      query: () => "/users/subscribers",
      transformResponse: (response: unknown) => subscriptionsResponseSchema.parse(response),
      providesTags: ["Subscriptions"],
    }),

    getSubscribersByUserId: builder.query<SubscriptionsResponse, string>({
      query: (id) => `/users/subscribers/${id}`,
      transformResponse: (response: unknown) => subscriptionsResponseSchema.parse(response),
      providesTags: (_result, _error, id) => [{type: "Subscriptions", id: `subscribers:${id}`}],
    }),

    subscribeToUser: builder.mutation<void, SubscribeRequest>({
      query: (body) => ({
        url: "/users/subscriptions",
        method: "POST",
        body,
      }),
      invalidatesTags: ["Subscriptions"],
    }),

    unsubscribeFromUser: builder.mutation<void, SubscribeRequest>({
      query: (body) => ({
        url: "/users/subscriptions",
        method: "DELETE",
        body,
      }),
      invalidatesTags: ["Subscriptions"],
    }),

    getAdminPlatformStatistics: builder.query<AdminUsersPlatformStatistics, void>({
      query: () => "/users/admin/statistics/platform",
      transformResponse: (response: unknown) => adminUsersPlatformStatisticsSchema.parse(response),
    }),

    getAdminUserStatistics: builder.query<AdminUsersUserStatistics, string>({
      query: (userId) => `/users/admin/users/${userId}/statistics`,
      transformResponse: (response: unknown) => adminUsersUserStatisticsSchema.parse(response),
    }),
  }),
});

export default usersApi;
