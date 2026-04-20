import { z } from "zod";
import {
  dealStatsSchema,
  myStatisticsSchema,
  offerStatsSchema,
  reportOnMyOffersStatsSchema,
  reportStatsSchema,
  reviewStatsSchema,
} from "./schemas";

export type DealStats = z.infer<typeof dealStatsSchema>;
export type OfferStats = z.infer<typeof offerStatsSchema>;
export type ReviewStats = z.infer<typeof reviewStatsSchema>;
export type ReportOnMyOffersStats = z.infer<typeof reportOnMyOffersStatsSchema>;
export type ReportStats = z.infer<typeof reportStatsSchema>;
export type MyStatistics = z.infer<typeof myStatisticsSchema>;
