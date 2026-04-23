import {z} from "zod";
import {
  favoriteOffersCursorSchema,
  favoritedOfferSchema,
  getFavoriteOffersResponseSchema,
  getOffersResponseSchema,
  listTagsResponseSchema,
  listOfferReportsResponseSchema,
  offerActionSchema,
  offerReportDetailsSchema,
  offerReportMessageSchema,
  offerReportSchema,
  offerReportStatusSchema,
  offerReportThreadSchema,
  offerReportsForOfferSchema,
  offerSchema,
  offerTypeSchema,
  universalCursorSchema,
} from "@/features/offers/model/schemas.ts";

export type OfferAction = z.Infer<typeof offerActionSchema>;
export type OfferType = z.Infer<typeof offerTypeSchema>;
export type OfferReportStatus = z.Infer<typeof offerReportStatusSchema>;

export type Offer = z.Infer<typeof offerSchema>;
export type OfferReport = z.Infer<typeof offerReportSchema>;
export type OfferReportMessage = z.Infer<typeof offerReportMessageSchema>;
export type OfferReportThread = z.Infer<typeof offerReportThreadSchema>;
export type OfferReportsForOffer = z.Infer<typeof offerReportsForOfferSchema>;
export type OfferReportDetails = z.Infer<typeof offerReportDetailsSchema>;
export type UniversalCursor = z.Infer<typeof universalCursorSchema>;
export type FavoriteOffersCursor = z.Infer<typeof favoriteOffersCursorSchema>;
export type FavoritedOffer = z.Infer<typeof favoritedOfferSchema>;
export type ListOfferReportsResponse = z.Infer<typeof listOfferReportsResponseSchema>;

export type SortType = "ByTime" | "ByPopularity";

export interface OffersListParams {
  sort: SortType;
  cursor_created_at?: string;
  cursor_views?: number;
  cursor_id?: string;
  cursor_limit?: number;
}

export interface GetOffersParams extends OffersListParams {
  my?: boolean;
  tags?: string[];
  withoutTags?: boolean;
}

export type GetSubscribedOffersParams = OffersListParams;

export type GetOffersResponse = z.Infer<typeof getOffersResponseSchema>;
export type GetFavoriteOffersResponse = z.Infer<typeof getFavoriteOffersResponseSchema>;
export type ListTagsResponse = z.Infer<typeof listTagsResponseSchema>;

export interface GetFavoriteOffersParams {
  cursor_favorited_at?: string;
  cursor_id?: string;
  cursor_limit?: number;
}

export interface CreateOfferRequest {
  name: string;
  description: string;
  action: OfferAction;
  type: OfferType;
  tags?: string[];
  photos?: File[];
  latitude?: number;
  longitude?: number;
}

export interface UpdateOfferRequest {
  name?: string;
  description?: string;
  action?: OfferAction;
  type?: OfferType;
  tags?: string[];
  photos?: File[];
  deletePhotoIds?: string[];
  latitude?: number | null;
  longitude?: number | null;
}

export interface CreateOfferReportRequest {
  message: string;
}

export interface ResolveOfferReportRequest {
  accepted: boolean;
  comment?: string;
}
