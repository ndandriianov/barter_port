import {Link} from "react-router-dom";
import {useState} from "react";
import dealsApi from "@/features/deals/api/dealsApi.ts";

function DealsList() {
  const [myOnly, setMyOnly] = useState(false);
  const [openOnly, setOpenOnly] = useState(false);

  const {data, isLoading, isFetching, error, refetch} = dealsApi.useGetDealsQuery({
    my: myOnly || undefined,
    open: openOnly || undefined,
  });

  if (isLoading) return <div>Загрузка сделок...</div>;
  if (error) return <div>Не удалось загрузить сделки</div>;
  if (!data) return <div>Список сделок недоступен</div>;

  return (
    <section>
      <div>
        <label>
          <input
            type="checkbox"
            checked={myOnly}
            onChange={(e) => setMyOnly(e.target.checked)}
          />
          Только мои
        </label>
        <label>
          <input
            type="checkbox"
            checked={openOnly}
            onChange={(e) => setOpenOnly(e.target.checked)}
          />
          Только открытые
        </label>
        <button type="button" onClick={() => refetch()} disabled={isFetching}>
          Обновить
        </button>
      </div>

      {data.data.length === 0 ? (
        <div>Сделок пока нет</div>
      ) : (
        data.data.map((dealId) => (
          <div key={dealId}>
            <Link to={`/deals/${dealId}`}>{dealId}</Link>
          </div>
        ))
      )}
    </section>
  );
}

export default DealsList;

