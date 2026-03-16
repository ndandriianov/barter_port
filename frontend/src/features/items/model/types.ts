import {z} from "zod";
import {
  getItemsResponseSchema,
  itemActionSchema,
  itemSchema,
  itemTypeSchema,
  universalCursorSchema
} from "@/features/items/model/schemas.ts";

export type ItemAction = z.Infer<typeof itemActionSchema>
export type ItemType = z.Infer<typeof itemTypeSchema>;

export type Item = z.Infer<typeof itemSchema>;
export type UniversalCursor = z.Infer<typeof universalCursorSchema>

export type SortType = "ByTime" | "ByPopularity";

export interface GetItemsParams {
  sort_type: SortType;
  created_at?: string;
  views?: number;
  id?: string;
  limit?: number;
}

export type GetItemsResponse = z.Infer<typeof getItemsResponseSchema>;

export interface CreateItemRequest {
  name: string;
  description: string;
  action: ItemAction;
  type: ItemType;
}

