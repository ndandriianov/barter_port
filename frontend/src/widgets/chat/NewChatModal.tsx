import { useMemo, useState } from "react";
import {
  Alert,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  MenuItem,
  TextField,
} from "@mui/material";
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
    <Dialog open onClose={onClose} fullWidth maxWidth="xs">
      <DialogTitle>Новый чат</DialogTitle>
      <DialogContent sx={{ display: "flex", flexDirection: "column", gap: 2, pt: 1 }}>
        <TextField
          select
          label="Пользователь"
          value={selected}
          onChange={(e) => {
            if (error) {
              reset();
            }
            setSelected(e.target.value);
          }}
          disabled={isLoading}
        >
          <MenuItem value="">Выберите пользователя</MenuItem>
          {users.map((u) => (
            <MenuItem key={u.id} value={u.id}>
              {u.name || u.id}
            </MenuItem>
          ))}
        </TextField>

        {error && <Alert severity="error">{getCreateChatErrorMessage(error)}</Alert>}

        {existingDirectChat && (
          <Alert severity="info">
            Чат с этим пользователем уже существует. Будет открыт существующий диалог.
          </Alert>
        )}
      </DialogContent>
      <DialogActions>
        <Button
          onClick={() => {
            reset();
            onClose();
          }}
        >
          Отмена
        </Button>
        <Button onClick={handleCreate} disabled={!selected || isCreating} variant="contained">
          {isCreating ? "Создание..." : existingDirectChat ? "Открыть чат" : "Создать"}
        </Button>
      </DialogActions>
    </Dialog>
  );
}

export default NewChatModal;
