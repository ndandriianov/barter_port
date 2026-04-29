import {z} from "zod";
import {
  adminAuthPlatformStatisticsSchema,
  adminAuthUserStatisticsSchema,
} from "@/features/auth/model/schemas.ts";

export type AdminAuthPlatformStatistics = z.Infer<typeof adminAuthPlatformStatisticsSchema>;
export type AdminAuthUserStatistics = z.Infer<typeof adminAuthUserStatisticsSchema>;
