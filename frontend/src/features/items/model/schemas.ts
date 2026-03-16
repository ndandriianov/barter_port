import {z} from "zod";

export const itemTypeSchema = z.enum(["good", "service"]);
export const itemActionSchema = z.enum(["give", "take"]);

export const itemSchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string(),
  action: itemActionSchema,
  type: itemTypeSchema,
  views: z.number(),
  createdAt: z.string(),
})

export const universalCursorSchema = z.object({
  id: z.string(),
  createdAt: z.string().nullable().optional(),
  views: z.number().nullable().optional(),
})

export const getItemsResponseSchema = z.object({
  items: z.array(itemSchema),
  nextCursor: z.nullable(universalCursorSchema),
})
