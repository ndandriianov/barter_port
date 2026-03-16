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
  createdAt: z.date(),
  views: z.number(),
})

export const getItemsResponseSchema = z.object({
  items: z.array(itemSchema),
  cursor: z.nullable(universalCursorSchema),
})