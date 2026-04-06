import { useState } from "react";
import ChatList from "@/widgets/chat/ChatList.tsx";
import ChatWindow from "@/widgets/chat/ChatWindow.tsx";
import NewChatModal from "@/widgets/chat/NewChatModal.tsx";
import chatsApi from "@/features/chats/api/chatsApi.ts";
import type { Chat } from "@/features/chats/model/types.ts";

function ChatsPage() {
  const [selectedChatId, setSelectedChatId] = useState<string | null>(null);
  const [showNewChatModal, setShowNewChatModal] = useState(false);

  const { data: chats = [] } = chatsApi.useListChatsQuery();
  const selectedChat: Chat | undefined = chats.find((c) => c.id === selectedChatId);

  function handleChatCreated(chatId: string) {
    setShowNewChatModal(false);
    setSelectedChatId(chatId);
  }

  return (
    <div style={{ display: "flex", height: "calc(100vh - 60px)", overflow: "hidden" }}>
      <ChatList
        selectedChatId={selectedChatId}
        onSelect={setSelectedChatId}
        onNewChat={() => setShowNewChatModal(true)}
      />

      <div style={{ flex: 1, display: "flex", flexDirection: "column" }}>
        {selectedChat ? (
          <ChatWindow chatId={selectedChat.id} participants={selectedChat.participants} />
        ) : (
          <div style={{ flex: 1, display: "flex", alignItems: "center", justifyContent: "center", color: "#888" }}>
            Выберите чат или начните новый
          </div>
        )}
      </div>

      {showNewChatModal && (
        <NewChatModal
          onClose={() => setShowNewChatModal(false)}
          onCreated={handleChatCreated}
        />
      )}
    </div>
  );
}

export default ChatsPage;
