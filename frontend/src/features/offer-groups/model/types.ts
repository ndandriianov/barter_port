import { z } from "zod";
import {
  createOfferGroupDraftResponseSchema,
  createOfferGroupResponseSchema,
  listOfferGroupsResponseSchema,
  offerGroupOfferRefSchema,
  offerGroupSchema,
  offerGroupUnitSchema,
} from "@/features/offer-groups/model/schemas.ts";

export type OfferGroupOfferRef = z.Infer<typeof offerGroupOfferRefSchema>;
export type OfferGroupUnit = z.Infer<typeof offerGroupUnitSchema>;
export type OfferGroup = z.Infer<typeof offerGroupSchema>;
export type ListOfferGroupsResponse = z.Infer<typeof listOfferGroupsResponseSchema>;
export type CreateOfferGroupResponse = z.Infer<typeof createOfferGroupResponseSchema>;
export type CreateOfferGroupDraftResponse = z.Infer<typeof createOfferGroupDraftResponseSchema>;

export interface CreateOfferGroupRequest {
  name: string;
  description?: string;
  units: Array<{
    offers: OfferGroupOfferRef[];
  }>;
}

export interface CreateOfferGroupDraftRequest {
  selectedOfferIds: string[];
  name?: string;
  description?: string;
}
