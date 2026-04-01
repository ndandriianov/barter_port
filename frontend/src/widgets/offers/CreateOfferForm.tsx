import {useState} from "react";
import offersApi from "@/features/offers/api/offersApi";
import {useNavigate} from "react-router-dom";
import type {OfferAction, OfferType} from "@/features/offers/model/types";

function CreateOfferForm() {
  const [createOffer, {isLoading, error}] = offersApi.useCreateOfferMutation();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [action, setAction] = useState<OfferAction>("give");
  const [type, setType] = useState<OfferType>("good");
  const navigate = useNavigate();

  const submit = async (e: React.FormEvent) => {
	e.preventDefault();
	await createOffer({name, description, action, type}).unwrap();
	navigate("/offers");
  };

  return (
	<form onSubmit={submit}>
	  <input
		placeholder="Название"
		value={name}
		onChange={(e) => setName(e.target.value)}
	  />
	  <textarea
		placeholder="Описание"
		value={description}
		onChange={(e) => setDescription(e.target.value)}
	  />
	  <select value={action} onChange={(e) => setAction(e.target.value as OfferAction)}>
		<option value="give">Отдаю</option>
		<option value="take">Беру</option>
	  </select>
	  <select value={type} onChange={(e) => setType(e.target.value as OfferType)}>
		<option value="good">Товар</option>
		<option value="service">Услуга</option>
	  </select>
	  <button type="submit" disabled={isLoading}>
		Создать
	  </button>
	  {error && <div>Ошибка создания</div>}
	</form>
  );
}

export default CreateOfferForm;

