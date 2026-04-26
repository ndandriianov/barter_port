import { useEffect, useRef, useState } from "react";
import {
  Alert,
  Box,
  Button,
  Chip,
  Paper,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import SendOutlinedIcon from "@mui/icons-material/SendOutlined";
import chatsApi from "@/features/chats/api/chatsApi.ts";
import usersApi from "@/features/users/api/usersApi.ts";
import type { Message, User } from "@/features/chats/model/types.ts";
import { getErrorMessage } from "@/shared/utils/getErrorMessage.ts";

interface Props {
  chatId: string;
  participants: string[];
  readOnly?: boolean;
}

const POLL_INTERVAL_MS = 3000;

function ChatWindow({ chatId, participants, readOnly = false }: Props) {
  const [content, setContent] = useState("");
  const [sendError, setSendError] = useState<string | null>(null);
  const bottomRef = useRef<HTMLDivElement>(null);

  const { data: messages = [], refetch } = chatsApi.useGetMessagesQuery({ chatId });
  const { data: me } = usersApi.useGetCurrentUserQuery();
  const { data: allUsers = [] } = chatsApi.useListUsersQuery();
  const [sendMessage, { isLoading: isSending }] = chatsApi.useSendMessageMutation();

  // Map userId → name for participants
  const userMap = new Map<string, string>(
    allUsers
      .filter((u: User) => participants.includes(u.id))
      .map((u: User) => [u.id, u.name])
  );

  // Fallback: if a sender isn't in participants list (e.g. deal chat), still show short id
  function getSenderLabel(senderId: string): string {
    return userMap.get(senderId) ?? senderId.slice(0, 8);
  }

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "smooth" });
  }, [messages]);

  useEffect(() => {
    const interval = setInterval(() => {
      refetch();
    }, POLL_INTERVAL_MS);
    return () => clearInterval(interval);
  }, [refetch]);

  async function handleSend() {
    if (readOnly) return;
    const trimmed = content.trim();
    if (!trimmed) return;

    setSendError(null);

    try {
      await sendMessage({ chatId, body: { content: trimmed } }).unwrap();
      setContent("");
      await refetch();
    } catch (error) {
      setSendError(getErrorMessage(error) ?? "Не удалось отправить сообщение");
    }
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      if (!readOnly) {
        void handleSend();
      }
    }
  }

  return (
    <Paper
      variant="outlined"
      sx={{
        flex: 1,
        minHeight: 0,
        display: "flex",
        flexDirection: "column",
        borderRadius: 4,
        overflow: "hidden",
      }}
    >
      <Box px={3} py={2.5} borderBottom="1px solid" borderColor="divider">
        <Stack direction="row" justifyContent="space-between" alignItems="center" gap={2} flexWrap="wrap">
          <div>
            <Typography variant="h6" fontWeight={800}>
              Диалог
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {participants.length > 0 ? `${participants.length} участников` : "Состав участников уточняется"}
            </Typography>
          </div>
          {readOnly ? <Chip label="Только чтение" color="warning" /> : <Chip label="Активен" color="success" variant="outlined" />}
        </Stack>
      </Box>

      <Box sx={{ flex: 1, overflowY: "auto", p: 3, display: "flex", flexDirection: "column", gap: 1.5 }}>
        {messages.map((msg: Message) => {
          const isMe = me?.id === msg.sender_id;
          const senderLabel = getSenderLabel(msg.sender_id);

          return (
            <Box
              key={msg.id}
              sx={{
                display: "flex",
                flexDirection: "column",
                alignItems: isMe ? "flex-end" : "flex-start",
              }}
            >
              <Typography
                variant="caption"
                color="text.secondary"
                sx={{ mb: 0.5, px: 0.5 }}
              >
                {isMe ? "Вы" : senderLabel}
              </Typography>

              <Box
                sx={{
                  bgcolor: isMe ? "primary.main" : "background.default",
                  color: isMe ? "primary.contrastText" : "text.primary",
                  borderRadius: isMe ? "22px 22px 6px 22px" : "22px 22px 22px 6px",
                  px: 1.75,
                  py: 1.25,
                  maxWidth: { xs: "85%", md: "70%" },
                  wordBreak: "break-word",
                  boxShadow: isMe ? "0 12px 24px rgba(15,118,110,0.2)" : "none",
                }}
              >
                {msg.content}
              </Box>

              <Typography variant="caption" color="text.secondary" sx={{ mt: 0.5, px: 0.5 }}>
                {new Date(msg.created_at).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
              </Typography>
            </Box>
          );
        })}
        <div ref={bottomRef} />
      </Box>

      {readOnly ? (
        <Alert severity="warning" sx={{ borderRadius: 0 }}>
          Чат доступен только для просмотра.
        </Alert>
      ) : (
        <Box sx={{ borderTop: "1px solid", borderColor: "divider", p: 2.5, display: "flex", gap: 1.5, alignItems: "flex-end" }}>
          <TextField
            value={content}
            onChange={(e) => {
              setContent(e.target.value);
              if (sendError) {
                setSendError(null);
              }
            }}
            onKeyDown={handleKeyDown}
            placeholder="Сообщение... Enter отправляет, Shift+Enter переносит строку"
            multiline
            minRows={2}
            maxRows={5}
            fullWidth
            sx={{
              flex: 1,
            }}
          />
          <Button
            onClick={() => void handleSend()}
            disabled={!content.trim() || isSending}
            variant="contained"
            startIcon={<SendOutlinedIcon />}
            sx={{ minWidth: 132, alignSelf: "stretch" }}
          >
            {isSending ? "Отправка..." : "Отправить"}
          </Button>
        </Box>
      )}
      {sendError && (
        <Alert severity="error" sx={{ borderRadius: 0 }}>
          {sendError}
        </Alert>
      )}
    </Paper>
  );
}

export default ChatWindow;
