import { useMemo, useState } from "react";
import type { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import type { SerializedError } from "@reduxjs/toolkit";
import chatsApi from "@/features/chats/api/chatsApi.ts";
import { getErrorMessage } from "@/shared/utils/getErrorMessage.ts";
import { getStatusCode } from "@/shared/utils/getStatusCode.ts";

interface Props {
  onClose: () => void;
  onCreated: (chatId: string) => void;
}

function getCreateChatErrorMessage(
  error: FetchBaseQueryError | SerializedError | undefined,
): string | null {
  if (!error) {
    return null;
  }

  const backendMessage = getErrorMessage(error);
  const statusCode = getStatusCode(error);

  switch (statusCode) {
    case 400:
      return backendMessage ?? "Не удалось создать чат: передан некорректный пользователь.";
    case 401:
      return backendMessage ?? "Сессия истекла. Войдите снова и повторите попытку.";
    case 403:
      return backendMessage ?? "Чат можно создать только при взаимной подписке.";
    case 409:
      return backendMessage ?? "Чат с этим пользователем уже существует.";
    case 500:
      return backendMessage ?? "Не удалось создать чат из-за ошибки сервера.";
    default:
      return backendMessage ?? "Не удалось создать чат.";
  }
}

function NewChatModal({ onClose, onCreated }: Props) {
  const { data: users = [], isLoading } = chatsApi.useListUsersQuery();
  const { data: chats = [] } = chatsApi.useListChatsQuery();
  const [createChat, { isLoading: isCreating, error, reset }] = chatsApi.useCreateChatMutation();
  const [selected, setSelected] = useState<string>("");
  const existingDirectChat = useMemo(
    () => chats.find((chat) => !chat.deal_id && chat.participants.includes(selected)),
    [chats, selected],
  );

  async function handleCreate() {
    if (!selected) return;

    if (existingDirectChat) {
      reset();
      onCreated(existingDirectChat.id);
      return;
    }

    try {
      const chat = await createChat({ participant_id: selected }).unwrap();
      onCreated(chat.id);
    } catch {
      // Ошибка уже доступна в mutation state и отображается в UI.
    }
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
            onChange={(e) => {
              if (error) {
                reset();
              }
              setSelected(e.target.value);
            }}
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

        {error && (
          <div
            style={{
              marginBottom: 16,
              padding: "10px 12px",
              borderRadius: 4,
              border: "1px solid #f5c2c7",
              background: "#f8d7da",
              color: "#842029",
            }}
          >
            {getCreateChatErrorMessage(error)}
          </div>
        )}

        {existingDirectChat && (
          <div
            style={{
              marginBottom: 16,
              padding: "10px 12px",
              borderRadius: 4,
              border: "1px solid #b6d4fe",
              background: "#cfe2ff",
              color: "#084298",
            }}
          >
            Чат с этим пользователем уже существует. Будет открыт существующий чат.
          </div>
        )}

        <div style={{ display: "flex", gap: 8, justifyContent: "flex-end" }}>
          <button
            onClick={() => {
              reset();
              onClose();
            }}
            style={{ padding: "8px 16px", borderRadius: 4, border: "1px solid #ccc", cursor: "pointer", background: "#f5f5f5" }}
          >
            Отмена
          </button>
          <button
            onClick={handleCreate}
            disabled={!selected || isCreating}
            style={{ padding: "8px 16px", borderRadius: 4, border: "none", cursor: "pointer", background: "#1976d2", color: "#fff" }}
          >
            {isCreating ? "Создание..." : existingDirectChat ? "Открыть чат" : "Создать"}
          </button>
        </div>
      </div>
    </div>
  );
}

export default NewChatModal;
