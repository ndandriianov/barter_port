import {useParams} from "react-router-dom";
import dealsApi from "@/features/deals/api/dealsApi.ts";
import DraftCard from "@/widgets/deals/DraftCard.tsx";

function DraftPage() {
  const {draftId} = useParams<{ draftId: string }>();

  const {data, isLoading, error} = dealsApi.useGetDraftDealByIdQuery(draftId ?? "", {
    skip: !draftId,
  });
  const [confirmDraftDeal, {isLoading: isConfirming}] = dealsApi.useConfirmDraftDealMutation();
  const [cancelDraftDeal, {isLoading: isCancelling}] = dealsApi.useCancelDraftDealMutation();

  if (!draftId) return <div>Черновик не найден</div>;
  if (isLoading) return <div>Загрузка черновика...</div>;
  if (error) return <div>Не удалось загрузить черновик</div>;
  if (!data) return <div>Черновик не найден</div>;

  const onConfirm = async () => {
    await confirmDraftDeal(draftId);
  };

  const onCancel = async () => {
    await cancelDraftDeal(draftId);
  };

  return (
    <section>
      <h1>Черновой договор</h1>
      <DraftCard draft={data} />
      <div>
        <button type="button" onClick={onConfirm} disabled={isConfirming}>
          Подтвердить участие
        </button>
        <button type="button" onClick={onCancel} disabled={isCancelling}>
          Отменить участие
        </button>
      </div>
    </section>
  );
}

export default DraftPage;

