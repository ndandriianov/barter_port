import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "@/shared/api/baseApi.ts";
import {
  adminChatsPlatformStatisticsSchema,
  chatSchema,
  createChatRequestSchema,
  getMessagesResponseSchema,
  listChatsResponseSchema,
  listUsersResponseSchema,
  messageSchema,
  sendMessageRequestSchema,
} from "@/features/chats/model/schemas.ts";
import type {
  AdminChatsPlatformStatistics,
  Chat,
  CreateChatRequest,
  GetMessagesResponse,
  ListChatsResponse,
  ListUsersResponse,
  Message,
  SendMessageRequest,
} from "@/features/chats/model/types.ts";

const chatsApi = createApi({
  reducerPath: "chatsApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Chats", "Messages"],
  endpoints: (builder) => ({
    listChats: builder.query<ListChatsResponse, void>({
      query: () => "/chats",
      transformResponse: (response: unknown) => listChatsResponseSchema.parse(response),
      providesTags: ["Chats"],
    }),

    createChat: builder.mutation<Chat, CreateChatRequest>({
      query: (body) => ({
        url: "/chats",
        method: "POST",
        body: createChatRequestSchema.parse(body),
      }),
      transformResponse: (response: unknown) => chatSchema.parse(response),
      invalidatesTags: ["Chats"],
    }),

    getDealChat: builder.query<Chat, string>({
      query: (dealId) => `/chats/deals/${dealId}`,
      transformResponse: (response: unknown) => chatSchema.parse(response),
      providesTags: (_result, _error, dealId) => [{ type: "Chats", id: dealId }],
    }),

    getMessages: builder.query<GetMessagesResponse, { chatId: string; after?: string }>({
      query: ({ chatId, after }) => ({
        url: `/chats/${chatId}/messages`,
        params: after ? { after } : undefined,
      }),
      transformResponse: (response: unknown) => getMessagesResponseSchema.parse(response),
      providesTags: (_result, _error, { chatId }) => [{ type: "Messages", id: chatId }],
    }),

    sendMessage: builder.mutation<Message, { chatId: string; body: SendMessageRequest }>({
      query: ({ chatId, body }) => ({
        url: `/chats/${chatId}/messages`,
        method: "POST",
        body: sendMessageRequestSchema.parse(body),
      }),
      transformResponse: (response: unknown) => messageSchema.parse(response),
      invalidatesTags: (_result, _error, { chatId }) => [{ type: "Messages", id: chatId }],
    }),

    listUsers: builder.query<ListUsersResponse, void>({
      query: () => "/chats/users",
      transformResponse: (response: unknown) => listUsersResponseSchema.parse(response),
    }),

    getAdminPlatformStatistics: builder.query<AdminChatsPlatformStatistics, void>({
      query: () => "/chats/admin/statistics/platform",
      transformResponse: (response: unknown) => adminChatsPlatformStatisticsSchema.parse(response),
    }),
  }),
});

export default chatsApi;
