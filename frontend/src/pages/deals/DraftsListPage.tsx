import {Link} from "react-router-dom";
import DraftsList from "@/widgets/deals/DraftsList.tsx";

function DraftsListPage() {
  return (
    <section>
      <h1>Мои черновые договоры</h1>
      <Link to="/deals/drafts/create">Создать черновой договор</Link>
      <p>Нажмите на ID, чтобы открыть страницу деталей сделки.</p>
      <DraftsList />
    </section>
  );
}

export default DraftsListPage;
