import type {Offer, OfferAction, OfferType} from "@/features/offers/model/types.ts";

const actionLabels: Record<OfferAction, string> = {
  give: "Отдаю",
  take: "Ищу",
};

const typeLabels: Record<OfferType, string> = {
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

interface OfferCardProps {
  offer: Offer;
}

function OfferCard({offer}: OfferCardProps) {
  return (
	<article>
	  <h3>{offer.name}</h3>
	  <div>authorId: {offer.authorId}</div>
	  <div>{typeLabels[offer.type]} • {actionLabels[offer.action]}</div>
	  <p>{offer.description}</p>
	  <div>Просмотры: {offer.views}</div>
	  <div>Создано: {formatCreatedAt(offer.createdAt)}</div>
	</article>
  );
}

export default OfferCard;

