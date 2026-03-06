export enum ItemAction {
  Give = 0,
  Take = 1,
}

export enum ItemType {
  Good = 0,
  Service = 1,
}

export interface Item {
  id: string;
  name: string;
  description: string;
  action: ItemAction;
  type: ItemType;
  views: number;
  createdAt: string;
}

export interface UniversalCursor {
  id: string;
  createdAt: string;
  views: number;
}

export type SortType = "ByTime" | "ByPopularity";

export interface GetItemsParams {
  sort_type: SortType;
  created_at?: string;
  views?: number;
  id?: string;
  limit?: number;
}

export interface GetItemsResponse {
  items: Item[];
  cursor: UniversalCursor;
}

export type CreateItemAction = "give" | "take";
export type CreateItemType = "good" | "service";

export interface CreateItemRequest {
  name: string;
  description: string;
  action: CreateItemAction;
  type: CreateItemType;
}

