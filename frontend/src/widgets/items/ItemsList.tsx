import {useState} from "react";
import itemsApi from "@/features/items/api/itemsApi.ts";
import type {ItemAction, ItemType, SortType} from "@/features/items/model/types.ts";

const actionLabels: Record<ItemAction, string> = {
  give: "Отдаю",
  take: "Ищу",
};

const typeLabels: Record<ItemType, string> = {
  good: "Товар",
  service: "Услуга",
};

const formatCreatedAt = (value: string) =>
  new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  }).format(new Date(value));

function ItemsList() {
  const [sortType, setSortType] = useState<SortType>("ByTime");
  const {data, isLoading, isFetching, error, refetch} = itemsApi.useGetItemsQuery({
    sort: sortType,
    cursor_limit: 20,
  });

  if (isLoading) return <div>Загрузка объявлений...</div>;
  if (error) return <div>Не удалось загрузить список items</div>;
  if (!data) return <div>Список items недоступен</div>;

  return (
    <div>
      <div>
        <label>
          Сортировка
          <select
            value={sortType}
            onChange={(e) => setSortType(e.target.value as SortType)}
          >
            <option value="ByTime">Сначала новые</option>
            <option value="ByPopularity">По популярности</option>
          </select>
        </label>
        <button type="button" onClick={() => refetch()} disabled={isFetching}>
          Обновить
        </button>
      </div>

      {data.items.length === 0 ? (
        <div>Пока нет объявлений</div>
      ) : (
        data.items.map((item) => (
          <article key={item.id}>
            <h3>{item.name}</h3>
            <div>{typeLabels[item.type]} • {actionLabels[item.action]}</div>
            <p>{item.description}</p>
            <div>Просмотры: {item.views}</div>
            <div>Создано: {formatCreatedAt(item.createdAt)}</div>
          </article>
        ))
      )}
    </div>
  );
}

export default ItemsList;
