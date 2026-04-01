import {Link} from "react-router-dom";
import dealsApi from "@/features/deals/api/dealsApi.ts";

function DraftsList() {
  const {data, isLoading, error, refetch, isFetching} = dealsApi.useGetMyDraftDealsQuery({
    createdByMe: false,
    participating: true
  });

  if (isLoading) return <div>Загрузка черновиков...</div>;
  if (error) return <div>Не удалось загрузить черновики</div>;
  if (!data) return <div>Черновики недоступны</div>;

  return (
    <section>
      <button type="button" onClick={() => refetch()} disabled={isFetching}>
        Обновить
      </button>

      {data.length === 0 ? (
        <div>У вас пока нет черновых договоров</div>
      ) : (
        data.map((draftId) => (
          <div key={draftId}>
            <Link to={`/deals/drafts/${draftId}`}>{draftId}</Link>
          </div>
        ))
      )}
    </section>
  );
}

export default DraftsList;
