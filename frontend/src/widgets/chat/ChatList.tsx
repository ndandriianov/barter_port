import chatsApi from "@/features/chats/api/chatsApi.ts";
import type { Chat } from "@/features/chats/model/types.ts";

interface Props {
  selectedChatId: string | null;
  onSelect: (chatId: string) => void;
  onNewChat: () => void;
}

function ChatList({ selectedChatId, onSelect, onNewChat }: Props) {
  const { data: chats = [], isLoading } = chatsApi.useListChatsQuery();

  return (
    <div style={{ width: 260, borderRight: "1px solid #e0e0e0", display: "flex", flexDirection: "column", height: "100%" }}>
      <div style={{ padding: "12px 16px", borderBottom: "1px solid #e0e0e0", display: "flex", justifyContent: "space-between", alignItems: "center" }}>
        <strong>Чаты</strong>
        <button
          onClick={onNewChat}
          style={{ padding: "4px 12px", borderRadius: 4, border: "none", cursor: "pointer", background: "#1976d2", color: "#fff", fontSize: 13 }}
        >
          + Новый
        </button>
      </div>

      <div style={{ overflowY: "auto", flex: 1 }}>
        {isLoading && <p style={{ padding: 16, color: "#888" }}>Загрузка...</p>}
        {!isLoading && chats.length === 0 && (
          <p style={{ padding: 16, color: "#888", fontSize: 14 }}>Нет чатов</p>
        )}
        {chats.map((chat: Chat) => (
          <div
            key={chat.id}
            onClick={() => onSelect(chat.id)}
            style={{
              padding: "12px 16px",
              cursor: "pointer",
              background: selectedChatId === chat.id ? "#e3f2fd" : "transparent",
              borderBottom: "1px solid #f0f0f0",
            }}
          >
            <div style={{ fontSize: 13, fontWeight: 500 }}>
              {chat.deal_id ? `Сделка` : "Личный чат"}
            </div>
            <div style={{ fontSize: 11, color: "#888", marginTop: 2, overflow: "hidden", textOverflow: "ellipsis", whiteSpace: "nowrap" }}>
              {chat.id.slice(0, 8)}...
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}

export default ChatList;
