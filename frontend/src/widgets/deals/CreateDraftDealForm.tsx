import {useState} from "react";
import {useNavigate} from "react-router-dom";
import dealsApi from "@/features/deals/api/dealsApi.ts";
import type {OfferIDAndQuantity} from "@/features/deals/model/types.ts";

const initialOffer: OfferIDAndQuantity = {
  offerID: "",
  quantity: 1,
};

function CreateDraftDealForm() {
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [offers, setOffers] = useState<OfferIDAndQuantity[]>([initialOffer]);

  const [createDraftDeal, {isLoading, error}] = dealsApi.useCreateDraftDealMutation();
  const navigate = useNavigate();

  const updateOffer = (index: number, patch: Partial<OfferIDAndQuantity>) => {
    setOffers((prev) => prev.map((item, i) => (i === index ? {...item, ...patch} : item)));
  };

  const addOffer = () => {
    setOffers((prev) => [...prev, {...initialOffer}]);
  };

  const removeOffer = (index: number) => {
    setOffers((prev) => prev.filter((_, i) => i !== index));
  };

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();

    const normalizedOffers = offers
      .map((item) => ({
        offerID: item.offerID.trim(),
        quantity: Number(item.quantity),
      }))
      .filter((item) => item.offerID && item.quantity > 0);

    if (normalizedOffers.length === 0) return;

    const result = await createDraftDeal({
      name: name.trim() || undefined,
      description: description.trim() || undefined,
      offers: normalizedOffers,
    }).unwrap();

    navigate(`/deals/drafts/${result.id}`);
  };

  return (
    <form onSubmit={submit}>
      <h1>Создать черновой договор</h1>

      <input
        placeholder="Название (опционально)"
        value={name}
        onChange={(e) => setName(e.target.value)}
      />

      <textarea
        placeholder="Описание (опционально)"
        value={description}
        onChange={(e) => setDescription(e.target.value)}
      />

      <h2>Объявления</h2>
      {offers.map((offer, index) => (
        <div key={index}>
          <input
            placeholder="Offer ID"
            value={offer.offerID}
            onChange={(e) => updateOffer(index, {offerID: e.target.value})}
          />
          <input
            type="number"
            min={1}
            value={offer.quantity}
            onChange={(e) => updateOffer(index, {quantity: Number(e.target.value)})}
          />
          <button
            type="button"
            onClick={() => removeOffer(index)}
            disabled={offers.length === 1}
          >
            Удалить
          </button>
        </div>
      ))}

      <button type="button" onClick={addOffer}>
        Добавить объявление
      </button>

      <button type="submit" disabled={isLoading}>
        Создать черновик
      </button>

      {error && <div>Не удалось создать черновик</div>}
    </form>
  );
}

export default CreateDraftDealForm;

