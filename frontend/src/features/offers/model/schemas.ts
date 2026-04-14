import {z} from "zod";

export const offerTypeSchema = z.enum(["good", "service"]);
export const offerActionSchema = z.enum(["give", "take"]);

export const offerSchema = z.object({
  id: z.string(),
  authorId: z.string(),
  authorName: z.string().nullish(),
  name: z.string(),
  photoIds: z.array(z.string()).nullish().transform((value) => value ?? []),
  photoUrls: z.array(z.string()).nullish().transform((value) => value ?? []),
  description: z.string(),
  action: offerActionSchema,
  type: offerTypeSchema,
  views: z.number(),
  createdAt: z.string(),
  updatedAt: z.string().nullish(),
  isHidden: z.boolean().nullish().transform((value) => value ?? false),
  modificationBlocked: z.boolean().nullish().transform((value) => value ?? false),
});

export const offerReportStatusSchema = z.enum(["Pending", "Accepted", "Rejected"]);

export const offerReportSchema = z.object({
  id: z.string(),
  offerId: z.string(),
  offerAuthorId: z.string(),
  status: offerReportStatusSchema,
  createdAt: z.string(),
  reviewedAt: z.string().nullish(),
  reviewedBy: z.string().nullish(),
  resolutionComment: z.string().nullish(),
  appliedPenaltyDelta: z.number().nullish(),
});

export const offerReportMessageSchema = z.object({
  offerReportId: z.string(),
  authorId: z.string(),
  message: z.string(),
});

export const offerReportThreadSchema = z.object({
  report: offerReportSchema,
  messages: z.array(offerReportMessageSchema),
});

export const offerReportsForOfferSchema = z.object({
  offer: offerSchema,
  reports: z.array(offerReportThreadSchema),
});

export const offerReportDetailsSchema = z.object({
  report: offerReportSchema,
  offer: offerSchema,
  messages: z.array(offerReportMessageSchema),
});

export const listOfferReportsResponseSchema = z.array(offerReportSchema);

export const universalCursorSchema = z.object({
  id: z.string(),
  createdAt: z.string().nullable().optional(),
  views: z.number().nullable().optional(),
});

export const getOffersResponseSchema = z.object({
  offers: z.array(offerSchema),
  nextCursor: z.nullable(universalCursorSchema),
});
