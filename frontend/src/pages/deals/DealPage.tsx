import { Link, useParams } from "react-router-dom";
import { Alert, Box, Button, CircularProgress, Typography } from "@mui/material";
import { skipToken } from "@reduxjs/toolkit/query";
import dealsApi from "@/features/deals/api/dealsApi";
import chatsApi from "@/features/chats/api/chatsApi";
import usersApi from "@/features/users/api/usersApi.ts";
import DealCard from "@/widgets/deals/DealCard";
import ChatWindow from "@/widgets/chat/ChatWindow";
import { getStatusCode } from "@/shared/utils/getStatusCode";

function DealPage() {
  const { dealId } = useParams<{ dealId: string }>();

  const { data, isLoading, error } = dealsApi.useGetDealByIdQuery(dealId ?? "", {
    skip: !dealId,
    pollingInterval: 10_000,
  });

  const canShowDealChat = data?.status === "Discussion" || data?.status === "Confirmed";
  const {
    data: dealChat,
    isLoading: isDealChatLoading,
    error: dealChatError,
  } = chatsApi.useGetDealChatQuery(canShowDealChat && data ? data.id : skipToken);
  const { data: currentUser } = usersApi.useGetCurrentUserQuery();

  if (!dealId) return <Alert severity="warning">Сделка не найдена</Alert>;

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    const code = getStatusCode(error);
    if (code === 404) {
      return (
        <Box display="flex" flexDirection="column" alignItems="flex-start" gap={2}>
          <Alert severity="warning">Сделка не найдена</Alert>
          <Button component={Link} to="/deals" variant="outlined" size="small">
            К списку сделок
          </Button>
        </Box>
      );
    }
    if (code === 403) return <Alert severity="error">У вас нет доступа к этой сделке</Alert>;
    return <Alert severity="error">Не удалось загрузить сделку. Попробуйте позже</Alert>;
  }
  if (!data) return <Alert severity="warning">Сделка не найдена</Alert>;

  return (
    <Box maxWidth={700} mx="auto">
      <Typography variant="h4" fontWeight={700} mb={3}>
        Детали сделки
      </Typography>
      <DealCard deal={data} />

      {canShowDealChat && (
        <Box mt={3}>
          <Typography variant="h5" fontWeight={700} mb={2}>
            Чат сделки
          </Typography>

          {isDealChatLoading ? (
            <Box display="flex" justifyContent="center" py={4}>
              <CircularProgress />
            </Box>
          ) : getStatusCode(dealChatError) === 404 ? (
            <Alert severity="info">Чат этой сделки пока недоступен</Alert>
          ) : dealChatError ? (
            getStatusCode(dealChatError) === 403 ? (
              <Alert severity="warning">У вас нет доступа к чату этой сделки</Alert>
            ) : (
              <Alert severity="error">Не удалось загрузить чат сделки</Alert>
            )
          ) : !dealChat ? (
            <Alert severity="info">Чат этой сделки пока недоступен</Alert>
          ) : (
            <Box sx={{ border: "1px solid #e0e0e0", borderRadius: 2, height: 520, overflow: "hidden" }}>
              <ChatWindow
                chatId={dealChat.id}
                participants={dealChat.participants}
                readOnly={currentUser?.isAdmin === true}
              />
            </Box>
          )}
        </Box>
      )}
    </Box>
  );
}

export default DealPage;
