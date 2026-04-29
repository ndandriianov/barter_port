import {z} from "zod";

export const userSchema = z.object({
  id: z.string(),
  name: z.string().optional(),
  bio: z.string().optional(),
  avatarUrl: z.string().optional(),
  phoneNumber: z.string().optional(),
});

export const meSchema = userSchema.extend({
  email: z.string(),
  createdAt: z.string(),
  isAdmin: z.boolean(),
  reputationPoints: z.number(),
  currentLatitude: z.number().nullish().transform((v) => v ?? null),
  currentLongitude: z.number().nullish().transform((v) => v ?? null),
});

export const userAvatarUploadSchema = z.object({
  avatarUrl: z.string(),
});

export const reputationEventSchema = z.object({
  id: z.string().uuid(),
  sourceType: z.enum([
    "deals.offer_report.penalty",
    "deals.deal_failure.responsible",
    "deals.deal_completion.reward",
    "deals.review_creation.reward",
  ]),
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

export const adminUsersPlatformStatisticsSchema = z.object({
  reputation: z.object({
    average: z.number(),
    median: z.number(),
    topUsers: z.array(
      z.object({
        userId: z.string().uuid(),
        name: z.string().optional(),
        reputationPoints: z.number().int(),
      }),
    ),
  }),
});

export const adminUsersUserStatisticsSchema = z.object({
  reputation: z.object({
    currentPoints: z.number().int(),
    history: reputationEventsResponseSchema,
  }),
  social: z.object({
    followersCount: z.number().int(),
    subscriptionsCount: z.number().int(),
  }),
});
