import {z} from "zod";

export const itemTypeSchema = z.enum(["good", "service"]);
export const offerActionSchema = z.enum(["give", "take"]);
export const dealStatusSchema = z.enum([
  "LookingForParticipants",
  "Discussion",
  "Confirmed",
  "Completed",
  "Cancelled",
  "Failed",
]);

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

export const draftIdAndParticipantsSchema = z.object({
  id: z.string(),
  participants: z.array(z.string()),
});

export const getMyDraftDealsResponseSchema = z.array(draftIdAndParticipantsSchema);

export const itemSchema = z.object({
  id: z.string(),
  authorId: z.string(),
  providerId: z.string().optional(),
  receiverId: z.string().optional(),
  name: z.string(),
  description: z.string(),
  type: itemTypeSchema,
  updatedAt: z.string().optional(),
  quantity: z.number().int(),
});

export const dealSchema = z.object({
  id: z.string(),
  name: z.string().optional(),
  description: z.string().optional(),
  createdAt: z.string(),
  updatedAt: z.string().optional(),
  status: dealStatusSchema,
  items: z.array(itemSchema),
});

export const changeDealStatusRequestSchema = z.object({
  expectedStatus: dealStatusSchema,
});

export const updateDealItemRequestSchema = z.object({
  name: z.string().optional(),
  description: z.string().optional(),
  quantity: z.number().int().min(1).optional(),
  claimProvider: z.boolean().optional(),
  releaseProvider: z.boolean().optional(),
  claimReceiver: z.boolean().optional(),
  releaseReceiver: z.boolean().optional(),
});

export const dealIdAndParticipantsSchema = z.object({
  id: z.string(),
  status: dealStatusSchema,
  participants: z.array(z.string()),
});

export const getDealsResponseSchema = z.array(dealIdAndParticipantsSchema);
