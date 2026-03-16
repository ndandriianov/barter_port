import type {Item, ItemAction, ItemType} from "@/features/items/model/types.ts";

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

interface ItemCardProps {
  item: Item;
}

function ItemCard({item}: ItemCardProps) {
  return (
    <article>
      <h3>{item.name}</h3>
      <div>{typeLabels[item.type]} • {actionLabels[item.action]}</div>
      <p>{item.description}</p>
      <div>Просмотры: {item.views}</div>
      <div>Создано: {formatCreatedAt(item.createdAt)}</div>
    </article>
  );
}

export default ItemCard;
