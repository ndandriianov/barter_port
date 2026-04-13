import {z} from "zod";

export const offerTypeSchema = z.enum(["good", "service"]);
export const offerActionSchema = z.enum(["give", "take"]);

export const offerSchema = z.object({
  id: z.string(),
  authorId: z.string(),
  authorName: z.string().nullish(),
  name: z.string(),
  photoUrls: z.array(z.string()).nullish().transform((value) => value ?? []),
  description: z.string(),
  action: offerActionSchema,
  type: offerTypeSchema,
  views: z.number(),
  createdAt: z.string(),
  updatedAt: z.string().nullish(),
});

export const universalCursorSchema = z.object({
  id: z.string(),
  createdAt: z.string().nullable().optional(),
  views: z.number().nullable().optional(),
});

export const getOffersResponseSchema = z.object({
  offers: z.array(offerSchema),
  nextCursor: z.nullable(universalCursorSchema),
});
