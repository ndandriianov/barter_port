import { useMemo, useState } from "react";
import { Box, Button, ButtonGroup, Stack, Typography } from "@mui/material";
import ChatBubbleOutlineOutlinedIcon from "@mui/icons-material/ChatBubbleOutlineOutlined";
import ForumOutlinedIcon from "@mui/icons-material/ForumOutlined";
import HandshakeOutlinedIcon from "@mui/icons-material/HandshakeOutlined";
import { Link as RouterLink, useLocation } from "react-router-dom";
import ChatList from "@/widgets/chat/ChatList.tsx";
import ChatWindow from "@/widgets/chat/ChatWindow.tsx";
import NewChatModal from "@/widgets/chat/NewChatModal.tsx";
import chatsApi from "@/features/chats/api/chatsApi.ts";
import type { Chat } from "@/features/chats/model/types.ts";
import { appRoutes } from "@/shared/config/appRoutes.ts";

interface ChatsPageProps {
  defaultMode?: "all" | "direct" | "deal";
}

function ChatsPage({ defaultMode = "all" }: ChatsPageProps) {
  const location = useLocation();
  const [selectedChatId, setSelectedChatId] = useState<string | null>(null);
  const [showNewChatModal, setShowNewChatModal] = useState(false);

  const { data: chats = [] } = chatsApi.useListChatsQuery();
  const requestedChatId = typeof location.state?.chatId === "string" ? location.state.chatId : null;
  const visibleChats = useMemo(() => {
    if (defaultMode === "direct") {
      return chats.filter((chat) => !chat.deal_id);
    }

    if (defaultMode === "deal") {
      return chats.filter((chat) => Boolean(chat.deal_id));
    }

    return chats;
  }, [chats, defaultMode]);
  const activeChatId = useMemo(() => {
    if (selectedChatId && visibleChats.some((chat) => chat.id === selectedChatId)) {
      return selectedChatId;
    }

    if (requestedChatId && visibleChats.some((chat) => chat.id === requestedChatId)) {
      return requestedChatId;
    }

    return visibleChats[0]?.id ?? null;
  }, [requestedChatId, selectedChatId, visibleChats]);
  const selectedChat: Chat | undefined = chats.find((c) => c.id === activeChatId);

  function handleChatCreated(chatId: string) {
    setShowNewChatModal(false);
    setSelectedChatId(chatId);
  }

  return (
    <Stack spacing={3} sx={{ minHeight: "calc(100vh - 180px)" }}>
      <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2} flexWrap="wrap">
        <Box>
          <Typography variant="overline" color="primary.main">
            Сообщения / Все чаты
          </Typography>
          <Typography variant="h4" fontWeight={800} mb={1}>
            {defaultMode === "deal" ? "Чаты по сделкам" : defaultMode === "direct" ? "Личные сообщения" : "Все сообщения"}
          </Typography>
        </Box>
      </Box>

      <ButtonGroup
        variant="text"
        sx={{
          alignSelf: "flex-start",
          bgcolor: "background.paper",
          borderRadius: 999,
          p: 0.75,
          boxShadow: "0 10px 30px rgba(15, 23, 42, 0.08)",
        }}
      >
        <Button
          component={RouterLink}
          to={appRoutes.messages.home}
          variant={defaultMode === "all" ? "contained" : "text"}
          startIcon={<ChatBubbleOutlineOutlinedIcon />}
        >
          Все
        </Button>
        <Button
          component={RouterLink}
          to={appRoutes.messages.direct}
          variant={defaultMode === "direct" ? "contained" : "text"}
          startIcon={<ForumOutlinedIcon />}
        >
          Личные
        </Button>
        <Button
          component={RouterLink}
          to={appRoutes.messages.deal}
          variant={defaultMode === "deal" ? "contained" : "text"}
          startIcon={<HandshakeOutlinedIcon />}
        >
          По сделкам
        </Button>
      </ButtonGroup>

      <Box display="flex" gap={2.5} flexDirection={{ xs: "column", md: "row" }} minHeight={0} flex={1}>
        <ChatList
          chats={visibleChats}
          mode={defaultMode}
          selectedChatId={activeChatId}
          onSelect={setSelectedChatId}
          onNewChat={() => setShowNewChatModal(true)}
        />

        <Box sx={{ flex: 1, minWidth: 0, minHeight: 0, display: "flex" }}>
          {selectedChat ? (
            <ChatWindow key={selectedChat.id} chatId={selectedChat.id} participants={selectedChat.participants} />
          ) : (
            <Box
              sx={{
                flex: 1,
                borderRadius: 4,
                border: "1px dashed",
                borderColor: "divider",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                p: 4,
                color: "text.secondary",
              }}
            >
              Выберите чат или создайте новый диалог.
            </Box>
          )}
        </Box>
      </Box>

      {showNewChatModal && (
        <NewChatModal
          onClose={() => setShowNewChatModal(false)}
          onCreated={handleChatCreated}
        />
      )}
    </Stack>
  );
}

export default ChatsPage;
