import {useParams} from "react-router-dom";
import {useAppSelector} from "@/hooks/redux.ts";
import itemsApi from "@/features/items/api/itemsApi.ts";
import type {GetItemsResponse} from "@/features/items/model/types.ts";
import ItemCard from "@/widgets/items/ItemCard.tsx";

function isGetItemsResponse(data: unknown): data is GetItemsResponse {
  return (
    typeof data === "object" &&
    data !== null &&
    "items" in data &&
    Array.isArray(data.items)
  );
}

function ItemPage() {
  const {itemId} = useParams<{itemId: string}>();
  const item = useAppSelector((state) => {
    const queries = state[itemsApi.reducerPath].queries;

    for (const queryState of Object.values(queries)) {
      if (queryState?.endpointName !== "getItems" || !isGetItemsResponse(queryState.data)) {
        continue;
      }

      const match = queryState.data.items.find((entry) => entry.id === itemId);

      if (match) return match;
    }

    return null;
  });

  if (!itemId) return <div>Объявление не найдено</div>;
  if (!item) return <div>Объявление не найдено в кеше RTK</div>;

  return (
    <section>
      <h1>{item.name}</h1>
      <ItemCard item={item} />
      <div>
        <button type="button">Откликнуться</button>
        <button type="button">Пожаловаться</button>
      </div>
    </section>
  );
}

export default ItemPage;
