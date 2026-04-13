import {z} from "zod";
import {meSchema, userSchema} from "@/features/users/model/schemas.ts";

export type User = z.Infer<typeof userSchema>;
export type Me = z.Infer<typeof meSchema>;

export interface UpdateCurrentUserRequest {
  name?: string;
  bio?: string;
  avatarUrl?: string;
}

export interface UploadCurrentUserAvatarResponse {
  avatarUrl: string;
}
