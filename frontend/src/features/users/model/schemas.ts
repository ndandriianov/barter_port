import {z} from "zod";

export const userSchema = z.object({
  id: z.string(),
  name: z.string().optional(),
  bio: z.string().optional(),
  avatarUrl: z.string().optional(),
});

export const meSchema = userSchema.extend({
  email: z.string(),
  createdAt: z.string(),
  isAdmin: z.boolean(),
  reputationPoints: z.number(),
});

export const userAvatarUploadSchema = z.object({
  avatarUrl: z.string(),
});

export const reputationEventSchema = z.object({
  id: z.string().uuid(),
  sourceType: z.string(),
  sourceId: z.string().uuid(),
  delta: z.number(),
  createdAt: z.string(),
  comment: z.string().optional(),
});

export const reputationEventsResponseSchema = z.array(reputationEventSchema);

export const subscribeRequestSchema = z.object({
  targetUserId: z.string().uuid(),
});

export const subscriptionsResponseSchema = z.array(userSchema);

