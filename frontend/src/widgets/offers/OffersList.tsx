import {useState} from "react";
import {Link} from "react-router-dom";
import offersApi from "@/features/offers/api/offersApi.ts";
import type {SortType} from "@/features/offers/model/types.ts";
import OfferCard from "@/widgets/offers/OfferCard.tsx";

function OffersList() {
  const [sortType, setSortType] = useState<SortType>("ByTime");
  const {data, isLoading, isFetching, error, refetch} = offersApi.useGetOffersQuery({
	sort: sortType,
	cursor_limit: 20,
  });

  if (isLoading) return <div>Загрузка объявлений...</div>;
  if (error) return <div>Не удалось загрузить список offers</div>;
  if (!data) return <div>Список offers недоступен</div>;

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

	  {data.offers.length === 0 ? (
		<div>Пока нет объявлений</div>
	  ) : (
		data.offers.map((offer) => (
		  <Link key={offer.id} to={`/offers/${offer.id}`}>
			<OfferCard offer={offer} />
		  </Link>
		))
	  )}
	</div>
  );
}

export default OffersList;

