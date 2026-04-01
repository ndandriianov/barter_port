import {useParams} from "react-router-dom";
import {useAppSelector} from "@/hooks/redux.ts";
import offersApi from "@/features/offers/api/offersApi.ts";
import type {GetOffersResponse, Offer} from "@/features/offers/model/types.ts";
import OfferCard from "@/widgets/offers/OfferCard.tsx";

interface CachedQueryState {
  endpointName?: string;
  data?: unknown;
}

function isGetOffersResponse(data: unknown): data is GetOffersResponse {
  return (
	typeof data === "object" &&
	data !== null &&
	"offers" in data &&
	Array.isArray(data.offers)
  );
}

function OfferPage() {
  const {offerId} = useParams<{offerId: string}>();
  const offer = useAppSelector((state) => {
	const queries = state[offersApi.reducerPath].queries;
	const queryStates = Object.values(queries) as CachedQueryState[];

	for (const queryState of queryStates) {
	  if (queryState?.endpointName !== "getOffers" || !isGetOffersResponse(queryState.data)) {
		continue;
	  }

	  const match = queryState.data.offers.find((entry: Offer) => entry.id === offerId);

	  if (match) return match;
	}

	return null;
  });

  if (!offerId) return <div>Объявление не найдено</div>;
  if (!offer) return <div>Объявление не найдено в кеше RTK</div>;

  return (
	<section>
	  <h1>{offer.name}</h1>
	  <OfferCard offer={offer} />
	  <div>
		<button type="button">Откликнуться</button>
		<button type="button">Пожаловаться</button>
	  </div>
	</section>
  );
}

export default OfferPage;


