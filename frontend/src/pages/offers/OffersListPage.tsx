import {Link} from "react-router-dom";
import OffersList from "@/widgets/offers/OffersList.tsx";

function OffersListPage() {
  return (
	<section>
	  <h1>Список offers</h1>
	  <Link to="/offers/create">Создать offer</Link>
	  <OffersList />
	</section>
  );
}

export default OffersListPage;

