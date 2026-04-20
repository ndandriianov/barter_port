import { z } from "zod";

export const dealStatsSchema = z.object({
  completed: z.number(),
  failed: z.number(),
  active: z.number(),
});

export const offerStatsSchema = z.object({
  total: z.number(),
  totalViews: z.number(),
});

export const reviewStatsSchema = z.object({
  written: z.number(),
  received: z.number(),
  averageRatingReceived: z.number().nullable(),
});

export const reportOnMyOffersStatsSchema = z.object({
  total: z.number(),
  pending: z.number(),
  accepted: z.number(),
  rejected: z.number(),
});

export const reportStatsSchema = z.object({
  onMyOffers: reportOnMyOffersStatsSchema,
  filedByMe: z.number(),
});

export const myStatisticsSchema = z.object({
  deals: dealStatsSchema,
  offers: offerStatsSchema,
  reviews: reviewStatsSchema,
  reports: reportStatsSchema,
});
