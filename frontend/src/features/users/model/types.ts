import {z} from "zod";
import {meSchema, reputationEventSchema, subscriptionsResponseSchema, userSchema} from "@/features/users/model/schemas.ts";

export type User = z.Infer<typeof userSchema>;
export type Me = z.Infer<typeof meSchema>;
export type ReputationEvent = z.Infer<typeof reputationEventSchema>;

export interface UpdateCurrentUserRequest {
  name?: string;
  bio?: string;
  avatarUrl?: string;
  phoneNumber?: string;
  currentLatitude?: number | null;
  currentLongitude?: number | null;
}

export interface UploadCurrentUserAvatarResponse {
  avatarUrl: string;
}

export interface SubscribeRequest {
  targetUserId: string;
}

export type SubscriptionsResponse = z.Infer<typeof subscriptionsResponseSchema>;
