import {useState} from "react";
import {useParams} from "react-router-dom";
import {useAppSelector} from "@/hooks/redux.ts";
import offersApi from "@/features/offers/api/offersApi.ts";
import authApi from "@/features/auth/api/authApi.ts";
import type {GetOffersResponse, Offer} from "@/features/offers/model/types.ts";
import OfferCard from "@/widgets/offers/OfferCard.tsx";
import RespondToOfferModal from "@/widgets/offers/RespondToOfferModal.tsx";

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
  const [isRespondModalOpen, setIsRespondModalOpen] = useState(false);
  const {data: meData} = authApi.useMeQuery();

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

  const canRespond = !!meData && offer.authorId !== meData.userId;

  return (
	<section>
	  <h1>{offer.name}</h1>
	  <OfferCard offer={offer} />
	  <div>
		{canRespond && (
		  <button type="button" onClick={() => setIsRespondModalOpen(true)}>
			Откликнуться
		  </button>
		)}
		<button type="button">Пожаловаться</button>
	  </div>

	  <RespondToOfferModal
		targetOffer={offer}
		isOpen={isRespondModalOpen}
		onClose={() => setIsRespondModalOpen(false)}
	  />
	</section>
  );
}

export default OfferPage;
