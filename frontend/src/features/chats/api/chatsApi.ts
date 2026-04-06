import { createApi } from "@reduxjs/toolkit/query/react";
import { baseQueryWithReauth } from "@/shared/api/baseApi.ts";
import type { Chat, Message, UserInfo, CreateChatRequest, SendMessageRequest } from "@/features/chats/model/types.ts";

const chatsApi = createApi({
  reducerPath: "chatsApi",
  baseQuery: baseQueryWithReauth,
  tagTypes: ["Chats", "Messages"],
  endpoints: (builder) => ({
    listChats: builder.query<Chat[], void>({
      query: () => "/chats",
      providesTags: ["Chats"],
    }),

    createChat: builder.mutation<Chat, CreateChatRequest>({
      query: (body) => ({
        url: "/chats",
        method: "POST",
        body,
      }),
      invalidatesTags: ["Chats"],
    }),

    getMessages: builder.query<Message[], { chatId: string; after?: string }>({
      query: ({ chatId, after }) => ({
        url: `/chats/${chatId}/messages`,
        params: after ? { after } : undefined,
      }),
      providesTags: (_result, _error, { chatId }) => [{ type: "Messages", id: chatId }],
    }),

    sendMessage: builder.mutation<Message, { chatId: string; body: SendMessageRequest }>({
      query: ({ chatId, body }) => ({
        url: `/chats/${chatId}/messages`,
        method: "POST",
        body,
      }),
      invalidatesTags: (_result, _error, { chatId }) => [{ type: "Messages", id: chatId }],
    }),

    listUsers: builder.query<UserInfo[], void>({
      query: () => "/chats/users",
    }),
  }),
});

export default chatsApi;
