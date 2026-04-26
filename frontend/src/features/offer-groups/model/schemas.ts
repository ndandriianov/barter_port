import { z } from "zod";
import { offerSchema } from "@/features/offers/model/schemas.ts";
import { createDraftDealResponseSchema } from "@/features/deals/model/schemas.ts";

export const offerGroupOfferRefSchema = z.object({
  offerId: z.string(),
});

export const offerGroupUnitSchema = z.object({
  id: z.string(),
  offers: z.array(offerSchema),
});

export const offerGroupSchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string().nullish(),
  draftDealsCount: z.number().int().nonnegative().nullish(),
  units: z.array(offerGroupUnitSchema),
});

export const listOfferGroupsResponseSchema = z.array(offerGroupSchema);

export const createOfferGroupResponseSchema = offerGroupSchema;

export const createOfferGroupDraftResponseSchema = createDraftDealResponseSchema;
