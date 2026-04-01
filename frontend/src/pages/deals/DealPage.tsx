import {useParams} from "react-router-dom";
import dealsApi from "@/features/deals/api/dealsApi.ts";
import DealCard from "@/widgets/deals/DealCard.tsx";

function DealPage() {
  const {dealId} = useParams<{ dealId: string }>();

  const {data, isLoading, error} = dealsApi.useGetDealByIdQuery(dealId ?? "", {
    skip: !dealId,
  });

  if (!dealId) return <div>Сделка не найдена</div>;
  if (isLoading) return <div>Загрузка сделки...</div>;
  if (error) return <div>Не удалось загрузить сделку</div>;
  if (!data) return <div>Сделка не найдена</div>;

  return (
    <section>
      <h1>Детали сделки</h1>
      <DealCard deal={data} />
    </section>
  );
}

export default DealPage;

