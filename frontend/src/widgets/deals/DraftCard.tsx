import type {Draft} from "@/features/deals/model/types.ts";

const formatDateTime = (value: string) =>
  new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  }).format(new Date(value));

interface DraftCardProps {
  draft: Draft;
}

function DraftCard({draft}: DraftCardProps) {
  return (
    <article>
      <h2>{draft.name ?? "Черновик сделки"}</h2>
      {draft.description && <p>{draft.description}</p>}
      <div>id: {draft.id}</div>
      <div>authorId: {draft.authorId}</div>
      <div>Создан: {formatDateTime(draft.createdAt)}</div>
      {draft.updatedAt && <div>Обновлен: {formatDateTime(draft.updatedAt)}</div>}

      <h3>Объявления в черновике</h3>
      {draft.offers.length === 0 ? (
        <div>Пусто</div>
      ) : (
        draft.offers.map((offer) => (
          <div key={offer.id}>
            <div>{offer.name}</div>
            <div>offerId: {offer.id}</div>
            <div>Количество: {offer.quantity}</div>
            {typeof offer.confirmed === "boolean" && (
              <div>Подтверждено: {offer.confirmed ? "да" : "нет"}</div>
            )}
          </div>
        ))
      )}
    </article>
  );
}

export default DraftCard;

