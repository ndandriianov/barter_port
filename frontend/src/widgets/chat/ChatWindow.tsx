import { useEffect, useRef, useState } from "react";
import chatsApi from "@/features/chats/api/chatsApi.ts";
import type { Message } from "@/features/chats/model/types.ts";

interface Props {
  chatId: string;
}

const POLL_INTERVAL_MS = 3000;

function ChatWindow({ chatId }: Props) {
  const [content, setContent] = useState("");
  const bottomRef = useRef<HTMLDivElement>(null);

  const { data: messages = [], refetch } = chatsApi.useGetMessagesQuery({ chatId });
  const [sendMessage] = chatsApi.useSendMessageMutation();

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
    const trimmed = content.trim();
    if (!trimmed) return;
    setContent("");
    await sendMessage({ chatId, body: { content: trimmed } });
    refetch();
  }

  function handleKeyDown(e: React.KeyboardEvent) {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  }

  return (
    <div style={{ flex: 1, display: "flex", flexDirection: "column", height: "100%" }}>
      <div style={{ flex: 1, overflowY: "auto", padding: 16, display: "flex", flexDirection: "column", gap: 8 }}>
        {messages.map((msg: Message) => (
          <div key={msg.id} style={{ display: "flex", flexDirection: "column", alignItems: "flex-start" }}>
            <div
              style={{
                background: "#f0f0f0",
                borderRadius: 12,
                padding: "8px 12px",
                maxWidth: "70%",
                wordBreak: "break-word",
              }}
            >
              {msg.content}
            </div>
            <div style={{ fontSize: 11, color: "#aaa", marginTop: 2, paddingLeft: 4 }}>
              {new Date(msg.created_at).toLocaleTimeString()}
            </div>
          </div>
        ))}
        <div ref={bottomRef} />
      </div>

      <div style={{ borderTop: "1px solid #e0e0e0", padding: 12, display: "flex", gap: 8 }}>
        <textarea
          value={content}
          onChange={(e) => setContent(e.target.value)}
          onKeyDown={handleKeyDown}
          placeholder="Сообщение..."
          rows={2}
          style={{ flex: 1, resize: "none", padding: 8, borderRadius: 4, border: "1px solid #ccc", fontFamily: "inherit", fontSize: 14 }}
        />
        <button
          onClick={handleSend}
          disabled={!content.trim()}
          style={{ padding: "8px 20px", borderRadius: 4, border: "none", cursor: "pointer", background: "#1976d2", color: "#fff", alignSelf: "flex-end" }}
        >
          Отправить
        </button>
      </div>
    </div>
  );
}

export default ChatWindow;
