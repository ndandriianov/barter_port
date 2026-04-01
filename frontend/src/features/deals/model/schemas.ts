import {z} from "zod";

export const itemTypeSchema = z.enum(["good", "service"]);
export const offerActionSchema = z.enum(["give", "take"]);

export const offerSchema = z.object({
  id: z.string(),
  authorId: z.string(),
  name: z.string(),
  description: z.string(),
  type: itemTypeSchema,
  action: offerActionSchema,
  views: z.number(),
  createdAt: z.string(),
});

export const offerInfoSchema = z.object({
  quantity: z.number().int(),
  confirmed: z.boolean().optional(),
});

export const offerWithInfoSchema = offerSchema.merge(offerInfoSchema);

export const draftSchema = z.object({
  id: z.string(),
  authorId: z.string(),
  name: z.string().optional(),
  description: z.string().optional(),
  createdAt: z.string(),
  updatedAt: z.string().optional(),
  offers: z.array(offerWithInfoSchema),
});

export const userConfirmSchema = z.object({
  userId: z.string(),
  confirmed: z.boolean(),
});

export const confirmDraftDealResponseSchema = z.object({
  users: z.array(userConfirmSchema),
});

export const createDraftDealResponseSchema = z.object({
  id: z.string(),
});

const idsListSchema = z.array(z.string());

export const getMyDraftDealsResponseSchema = idsListSchema;

export const itemSchema = z.object({
  id: z.string(),
  authorId: z.string(),
  providerId: z.string().optional(),
  receiverId: z.string().optional(),
  name: z.string(),
  description: z.string(),
  type: itemTypeSchema,
  updatedAt: z.string().optional(),
});

export const dealSchema = z.object({
  id: z.string(),
  name: z.string().optional(),
  description: z.string().optional(),
  createdAt: z.string(),
  updatedAt: z.string().optional(),
  items: z.array(itemSchema),
});

export const getDealsResponseSchema = z.object({
  data: z.array(z.string()),
});
