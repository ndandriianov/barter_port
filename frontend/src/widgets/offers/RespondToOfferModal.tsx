import {useMemo, useState} from "react";
import {useNavigate} from "react-router-dom";
import offersApi from "@/features/offers/api/offersApi.ts";
import dealsApi from "@/features/deals/api/dealsApi.ts";
import type {Offer} from "@/features/offers/model/types.ts";

interface RespondToOfferModalProps {
  targetOffer: Offer;
  isOpen: boolean;
  onClose: () => void;
}

function RespondToOfferModal({targetOffer, isOpen, onClose}: RespondToOfferModalProps) {
  const [selectedOfferId, setSelectedOfferId] = useState<string | null>(null);
  const navigate = useNavigate();

  const closeModal = () => {
    setSelectedOfferId(null);
    onClose();
  };

  const {data, isLoading, error} = offersApi.useGetOffersQuery(
    {
      sort: "ByTime",
      my: true,
      cursor_limit: 100,
    },
    {
      skip: !isOpen,
    },
  );

  const [createDraftDeal, {isLoading: isCreating, error: createError}] = dealsApi.useCreateDraftDealMutation();

  const selectedOffer = useMemo(
    () => data?.offers.find((entry) => entry.id === selectedOfferId),
    [data?.offers, selectedOfferId],
  );

  const submit = async () => {
    if (!selectedOffer) return;

    const result = await createDraftDeal({
      offers: [
        {offerID: targetOffer.id, quantity: 1},
        {offerID: selectedOffer.id, quantity: 1},
      ],
    }).unwrap();

    closeModal();
    navigate(`/deals/drafts/${result.id}`);
  };

  if (!isOpen) return null;

  return (
    <div
      role="dialog"
      aria-modal="true"
      style={{
        position: "fixed",
        inset: 0,
        backgroundColor: "rgba(0, 0, 0, 0.35)",
        display: "flex",
        alignItems: "center",
        justifyContent: "center",
        padding: 16,
      }}
    >
      <div style={{backgroundColor: "#fff", padding: 16, width: "min(680px, 100%)"}}>
        <h2>Выберите свое объявление для отклика</h2>
        <button type="button" onClick={closeModal}>
          Закрыть
        </button>

        {isLoading && <div>Загрузка ваших объявлений...</div>}
        {error && <div>Не удалось загрузить ваши объявления</div>}

        {!isLoading && !error && data && (
          <div>
            {data.offers.length === 0 ? (
              <div>У вас пока нет объявлений для отклика</div>
            ) : (
              data.offers.map((offer) => (
                <label key={offer.id} style={{display: "block", marginBottom: 8}}>
                  <input
                    type="radio"
                    name="selectedOffer"
                    value={offer.id}
                    checked={selectedOfferId === offer.id}
                    onChange={() => setSelectedOfferId(offer.id)}
                  />
                  {offer.name} - {offer.description}
                </label>
              ))
            )}
          </div>
        )}

        <button
          type="button"
          onClick={submit}
          disabled={!selectedOfferId || isCreating || isLoading || !!error}
        >
          Добавить к черновику
        </button>

        {createError && <div>Не удалось создать черновик сделки</div>}
      </div>
    </div>
  );
}

export default RespondToOfferModal;
