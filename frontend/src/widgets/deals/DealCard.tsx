import type {Deal} from "@/features/deals/model/types.ts";
const formatDateTime = (value: string) =>
  new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));
interface DealCardProps {
  deal: Deal;
}
function DealCard({deal}: DealCardProps) {
  return (
    <article>
      <h2>{deal.name ?? "Сделка"}</h2>
      {deal.description && <p>{deal.description}</p>}
      <div>id: {deal.id}</div>
      <div>Создана: {formatDateTime(deal.createdAt)}</div>
      {deal.updatedAt && <div>Обновлена: {formatDateTime(deal.updatedAt)}</div>}
      <h3>Позиции сделки</h3>
      {deal.items.length === 0 ? (
        <div>Позиции отсутствуют</div>
      ) : (
        deal.items.map((item) => (
          <div key={item.id}>
            <div>{item.name}</div>
            <div>{item.description}</div>
            <div>Тип: {item.type}</div>
            <div>authorId: {item.authorId}</div>
            {item.providerId && <div>providerId: {item.providerId}</div>}
            {item.receiverId && <div>receiverId: {item.receiverId}</div>}
          </div>
        ))
      )}
    </article>
  );
}
export default DealCard;
