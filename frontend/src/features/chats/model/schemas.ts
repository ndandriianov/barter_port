import {z} from "zod";

export const chatParticipantSchema = z.object({
  user_id: z.string(),
  user_name: z.string().nullish(),
});

export const chatSchema = z.object({
  id: z.string(),
  deal_id: z.string().nullable().optional(),
  participants: z.array(chatParticipantSchema),
  created_at: z.string(),
});

export const messageSchema = z.object({
  id: z.string(),
  chat_id: z.string(),
  sender_id: z.string(),
  content: z.string(),
  created_at: z.string(),
});

export const userSchema = z.object({
  id: z.string(),
  name: z.string(),
});

export const createChatRequestSchema = z.object({
  participant_id: z.string(),
});

export const sendMessageRequestSchema = z.object({
  content: z.string().min(1),
});

export const listChatsResponseSchema = z.array(chatSchema);

export const getMessagesResponseSchema = z.array(messageSchema);

export const listUsersResponseSchema = z.array(userSchema);
