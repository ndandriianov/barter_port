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
  name: z.string().optional(),
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
  participants: z.array(z.string()),
});

export const changeDealStatusRequestSchema = z.object({
  expectedStatus: dealStatusSchema,
});

export const addDealItemRequestSchema = z.object({
  offerId: z.string(),
  quantity: z.number().int().min(1),
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

export const dealJoinRequestSchema = z.object({
  userId: z.string(),
  dealId: z.string(),
  voters: z.array(z.string()),
});

export const getDealJoinRequestsResponseSchema = z.array(dealJoinRequestSchema);

export const getDealStatusVotesResponseItemSchema = z.object({
  userId: z.string(),
  vote: dealStatusSchema,
});

export const getDealStatusVotesResponseSchema = z.array(getDealStatusVotesResponseItemSchema);

export const dealIdAndParticipantsSchema = z.object({
  id: z.string(),
  name: z.string().optional(),
  status: dealStatusSchema,
  participants: z.array(z.string()),
});

export const getDealsResponseSchema = z.array(dealIdAndParticipantsSchema);

export const voteForFailureRequestSchema = z.object({
  userId: z.string(),
});

export const failureVoteSchema = z.object({
  userId: z.string(),
  vote: z.string(),
});

export const getFailureVotesResponseSchema = z.array(failureVoteSchema);

export const moderatorResolutionForFailureRequestSchema = z.object({
  confirmed: z.boolean(),
  userId: z.string().optional(),
  punishmentPoints: z.number().int().min(0).optional(),
  comment: z.string().optional(),
});

export const failureResolutionSchema = z.object({
  userId: z.string().optional(),
  confirmed: z.boolean().optional(),
  punishmentPoints: z.number().int().optional(),
  comment: z.string().optional(),
});

export const failureMaterialsSchema = z.object({
  deal: dealSchema,
  chatId: z.string().optional(),
});

export const failureModerationDealsResponseSchema = getDealsResponseSchema;
