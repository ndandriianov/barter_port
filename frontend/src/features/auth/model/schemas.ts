import {z} from "zod";

export const adminAuthPlatformStatisticsSchema = z.object({
  users: z.object({
    totalRegistered: z.number().int(),
    verifiedEmails: z.number().int(),
  }),
});

export const adminAuthUserStatisticsSchema = z.object({
  userId: z.string().uuid(),
  registeredAt: z.string(),
  emailVerified: z.boolean(),
});
