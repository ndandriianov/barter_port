import {useState} from "react";
import itemsApi from "@/features/items/api/itemsApi";
import {useNavigate} from "react-router-dom";
import type {ItemAction, ItemType} from "@/features/items/model/types";

function CreateItemForm() {
  const [createItem, {isLoading, error}] = itemsApi.useCreateItemMutation();
  const [name, setName] = useState("");
  const [description, setDescription] = useState("");
  const [action, setAction] = useState<ItemAction>("give");
  const [type, setType] = useState<ItemType>("good");
  const navigate = useNavigate();

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    await createItem({name, description, action, type}).unwrap();
    navigate("/");
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
      <select value={action} onChange={(e) => setAction(e.target.value as ItemAction)}>
        <option value="give">Отдаю</option>
        <option value="take">Беру</option>
      </select>
      <select value={type} onChange={(e) => setType(e.target.value as ItemType)}>
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

export default CreateItemForm;

