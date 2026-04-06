import { useState } from "react";
import chatsApi from "@/features/chats/api/chatsApi.ts";

interface Props {
  onClose: () => void;
  onCreated: (chatId: string) => void;
}

function NewChatModal({ onClose, onCreated }: Props) {
  const { data: users = [], isLoading } = chatsApi.useListUsersQuery();
  const [createChat] = chatsApi.useCreateChatMutation();
  const [selected, setSelected] = useState<string>("");

  async function handleCreate() {
    if (!selected) return;
    const chat = await createChat({ participant_id: selected }).unwrap();
    onCreated(chat.id);
  }

  return (
    <div style={{ position: "fixed", inset: 0, background: "rgba(0,0,0,0.4)", display: "flex", alignItems: "center", justifyContent: "center", zIndex: 100 }}>
      <div style={{ background: "#fff", borderRadius: 8, padding: 24, minWidth: 320 }}>
        <h3 style={{ marginTop: 0 }}>Новый чат</h3>

        {isLoading ? (
          <p>Загрузка...</p>
        ) : (
          <select
            value={selected}
            onChange={(e) => setSelected(e.target.value)}
            style={{ width: "100%", padding: "8px", marginBottom: 16, borderRadius: 4, border: "1px solid #ccc" }}
          >
            <option value="">Выберите пользователя</option>
            {users.map((u) => (
              <option key={u.id} value={u.id}>
                {u.name || u.id}
              </option>
            ))}
          </select>
        )}

        <div style={{ display: "flex", gap: 8, justifyContent: "flex-end" }}>
          <button onClick={onClose} style={{ padding: "8px 16px", borderRadius: 4, border: "1px solid #ccc", cursor: "pointer", background: "#f5f5f5" }}>
            Отмена
          </button>
          <button
            onClick={handleCreate}
            disabled={!selected}
            style={{ padding: "8px 16px", borderRadius: 4, border: "none", cursor: "pointer", background: "#1976d2", color: "#fff" }}
          >
            Создать
          </button>
        </div>
      </div>
    </div>
  );
}

export default NewChatModal;
