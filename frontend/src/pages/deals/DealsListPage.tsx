import {Link} from "react-router-dom";
import DealsList from "@/widgets/deals/DealsList.tsx";

function DealsListPage() {
  return (
    <section>
      <h1>Сделки</h1>
      <Link to="/deals/drafts">Мои черновики</Link>
      <div>
        <Link to="/deals/drafts/create">Создать черновой договор</Link>
      </div>
      <DealsList />
    </section>
  );
}

export default DealsListPage;

