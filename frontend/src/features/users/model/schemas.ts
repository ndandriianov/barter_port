import {z} from "zod";

export const userSchema = z.object({
  id: z.string(),
  name: z.string().optional(),
  bio: z.string().optional(),
});

export const meSchema = userSchema.extend({
  email: z.string(),
  createdAt: z.string(),
  isAdmin: z.boolean(),
});
