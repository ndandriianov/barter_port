import {Link} from "react-router-dom";
import ItemsList from "@/widgets/items/ItemsList.tsx";

function ItemsListPage() {
  return (
    <section>
      <h1>Список items</h1>
      <Link to="/items/create">Создать item</Link>
      <ItemsList />
    </section>
  );
}

export default ItemsListPage;
