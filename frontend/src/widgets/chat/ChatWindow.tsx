import { useEffect, useRef, useState } from "react";
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
    <div style={{ flex: 1, display: "flex", flexDirection: "column", height: "100%" }}>
      <div style={{ flex: 1, overflowY: "auto", padding: 16, display: "flex", flexDirection: "column", gap: 10 }}>
        {messages.map((msg: Message) => {
          const isMe = me?.id === msg.sender_id;
          const senderLabel = getSenderLabel(msg.sender_id);

          return (
            <div
              key={msg.id}
              style={{
                display: "flex",
                flexDirection: "column",
                alignItems: isMe ? "flex-end" : "flex-start",
              }}
            >
              {/* Имя отправителя */}
              <span style={{ fontSize: 11, color: "#888", marginBottom: 2, paddingLeft: isMe ? 0 : 4, paddingRight: isMe ? 4 : 0 }}>
                {isMe ? "Вы" : senderLabel}
              </span>

              {/* Пузырёк сообщения */}
              <div
                style={{
                  background: isMe ? "#1976d2" : "#f0f0f0",
                  color: isMe ? "#fff" : "#000",
                  borderRadius: isMe ? "16px 16px 4px 16px" : "16px 16px 16px 4px",
                  padding: "8px 12px",
                  maxWidth: "65%",
                  wordBreak: "break-word",
                  fontSize: 14,
                }}
              >
                {msg.content}
              </div>

              {/* Время */}
              <span style={{ fontSize: 11, color: "#aaa", marginTop: 2, paddingLeft: isMe ? 0 : 4, paddingRight: isMe ? 4 : 0 }}>
                {new Date(msg.created_at).toLocaleTimeString([], { hour: "2-digit", minute: "2-digit" })}
              </span>
            </div>
          );
        })}
        <div ref={bottomRef} />
      </div>

      {readOnly ? (
        <div
          style={{
            borderTop: "1px solid #e0e0e0",
            padding: 12,
            color: "#666",
            fontSize: 13,
            background: "#fafafa",
          }}
        >
          Чат доступен только для просмотра.
        </div>
      ) : (
        <div style={{ borderTop: "1px solid #e0e0e0", padding: 12, display: "flex", gap: 8 }}>
          <textarea
            value={content}
            onChange={(e) => {
              setContent(e.target.value);
              if (sendError) {
                setSendError(null);
              }
            }}
            onKeyDown={handleKeyDown}
            placeholder="Сообщение... (Enter — отправить)"
            rows={2}
            style={{
              flex: 1,
              resize: "none",
              padding: 8,
              borderRadius: 4,
              border: "1px solid #ccc",
              fontFamily: "inherit",
              fontSize: 14,
            }}
          />
          <button
            onClick={() => void handleSend()}
            disabled={!content.trim() || isSending}
            style={{
              padding: "8px 20px",
              borderRadius: 4,
              border: "none",
              cursor: !content.trim() || isSending ? "default" : "pointer",
              background: "#1976d2",
              color: "#fff",
              alignSelf: "flex-end",
              opacity: content.trim() && !isSending ? 1 : 0.5,
            }}
          >
            {isSending ? "Отправка..." : "Отправить"}
          </button>
        </div>
      )}
      {sendError && (
        <div
          style={{
            borderTop: readOnly ? undefined : "1px solid #f0d3d3",
            padding: "10px 12px",
            color: "#b42318",
            background: "#fef3f2",
            fontSize: 13,
          }}
        >
          {sendError}
        </div>
      )}
    </div>
  );
}

export default ChatWindow;
