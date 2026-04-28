import {
  Avatar,
  Box,
  Button,
  Chip,
  Divider,
  List,
  ListItemButton,
  ListItemText,
  Paper,
  Stack,
  Typography,
} from "@mui/material";
import AddOutlinedIcon from "@mui/icons-material/AddOutlined";
import ForumOutlinedIcon from "@mui/icons-material/ForumOutlined";
import HandshakeOutlinedIcon from "@mui/icons-material/HandshakeOutlined";
import type { Chat } from "@/features/chats/model/types.ts";
import { getUserDisplayName } from "@/shared/utils/getUserDisplayName.ts";

interface Props {
  chats: Chat[];
  mode: "all" | "direct" | "deal";
  selectedChatId: string | null;
  onSelect: (chatId: string) => void;
  onNewChat: () => void;
}

function ChatList({ chats, mode, selectedChatId, onSelect, onNewChat }: Props) {
  function getParticipantsLabel(chat: Chat): string {
    if (!chat.participants.length) return "Участники не указаны";

    return chat.participants
      .map((participant) => getUserDisplayName(participant.user_name, participant.user_id))
      .join(", ");
  }

  const personalChats = chats.filter((c) => !c.deal_id);
  const dealChats = chats.filter((c) => !!c.deal_id);

  const renderSection = (title: string, items: Chat[], kind: "direct" | "deal") => (
    <Stack spacing={1.25}>
      <Box display="flex" justifyContent="space-between" alignItems="center">
        <Typography variant="overline" color="text.secondary">
          {title}
        </Typography>
        <Chip size="small" label={items.length} />
      </Box>
      <List disablePadding sx={{ display: "flex", flexDirection: "column", gap: 1 }}>
        {items.map((chat) => (
          <Paper
            key={chat.id}
            variant="outlined"
            sx={{
              overflow: "hidden",
              borderColor: selectedChatId === chat.id ? "primary.main" : "divider",
              bgcolor: selectedChatId === chat.id ? "rgba(15,118,110,0.08)" : "background.paper",
            }}
          >
            <ListItemButton onClick={() => onSelect(chat.id)} sx={{ borderRadius: 3 }}>
              <Avatar
                sx={{
                  width: 40,
                  height: 40,
                  mr: 1.5,
                  bgcolor: kind === "deal" ? "secondary.main" : "primary.main",
                }}
              >
                {kind === "deal" ? <HandshakeOutlinedIcon fontSize="small" /> : <ForumOutlinedIcon fontSize="small" />}
              </Avatar>
              <ListItemText
                primary={kind === "deal" ? "Чат сделки" : "Личный чат"}
                secondary={getParticipantsLabel(chat)}
                primaryTypographyProps={{ fontWeight: 700 }}
                secondaryTypographyProps={{
                  sx: {
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                  },
                }}
              />
            </ListItemButton>
          </Paper>
        ))}
      </List>
    </Stack>
  );

  return (
    <Paper
      variant="outlined"
      sx={{
        width: { xs: "100%", md: 360 },
        minWidth: 0,
        display: "flex",
        flexDirection: "column",
        height: "100%",
        borderRadius: 4,
        overflow: "hidden",
      }}
    >
      <Box p={2.5}>
        <Stack direction="row" justifyContent="space-between" alignItems="center" gap={1.5}>
          <div>
            <Typography variant="h6" fontWeight={800}>
              Все чаты
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {mode === "deal" ? "Только переписки по сделкам" : mode === "direct" ? "Только личные диалоги" : "Личные и чаты по сделкам"}
            </Typography>
          </div>
          <Button onClick={onNewChat} variant="contained" size="small" startIcon={<AddOutlinedIcon />}>
            Новый
          </Button>
        </Stack>
      </Box>
      <Divider />

      <Box sx={{ p: 2, overflowY: "auto", flex: 1 }}>
        {chats.length === 0 ? (
          <Typography color="text.secondary" textAlign="center" py={6}>
            Чатов пока нет.
          </Typography>
        ) : (
          <Stack spacing={2}>
            {(mode === "all" || mode === "direct") && personalChats.length > 0 && renderSection("Личные", personalChats, "direct")}
            {(mode === "all" || mode === "deal") && dealChats.length > 0 && renderSection("По сделкам", dealChats, "deal")}
          </Stack>
        )}
      </Box>
    </Paper>
  );
}

export default ChatList;
