import { Link, useNavigate, useParams } from "react-router-dom";
import { Alert, Box, Button, CircularProgress, Typography } from "@mui/material";
import { skipToken } from "@reduxjs/toolkit/query";
import dealsApi from "@/features/deals/api/dealsApi";
import chatsApi from "@/features/chats/api/chatsApi";
import usersApi from "@/features/users/api/usersApi.ts";
import DealCard from "@/widgets/deals/DealCard";
import ChatWindow from "@/widgets/chat/ChatWindow";
import { getStatusCode } from "@/shared/utils/getStatusCode";
import type { DealStatus } from "@/features/deals/model/types.ts";

const FINAL_STATUSES: DealStatus[] = ["Completed", "Cancelled", "Failed"];

function DealPage() {
  const { dealId } = useParams<{ dealId: string }>();
  const navigate = useNavigate();

  const { data, isLoading, error } = dealsApi.useGetDealByIdQuery(dealId ?? "", {
    skip: !dealId,
    pollingInterval: 10_000,
  });

  const { data: currentUser } = usersApi.useGetCurrentUserQuery();

  const isFinalStatus = data ? FINAL_STATUSES.includes(data.status) : false;
  const isParticipant = data && currentUser ? data.participants.includes(currentUser.id) : false;
  const canAccessFailureResolution = Boolean(currentUser && (isParticipant || currentUser.isAdmin));

  const { data: failureResolution } = dealsApi.useGetModeratorResolutionForFailureQuery(
    canAccessFailureResolution && data ? data.id : skipToken,
    { pollingInterval: 10_000 },
  );
  const isFailurePending = failureResolution !== undefined && failureResolution.confirmed === undefined;

  const canShowDealChat = data
    ? data.status === "Discussion" ||
      data.status === "Confirmed" ||
      FINAL_STATUSES.includes(data.status)
    : false;

  const {
    data: dealChat,
    isLoading: isDealChatLoading,
    error: dealChatError,
  } = chatsApi.useGetDealChatQuery(canShowDealChat && data ? data.id : skipToken);

  const isChatReadOnly =
    currentUser?.isAdmin === true || isFinalStatus || isFailurePending;

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
      <Button
        size="small"
        variant="text"
        onClick={() => window.history.length > 1 ? navigate(-1) : navigate("/deals")}
        sx={{ mb: 2 }}
      >
        ← Назад
      </Button>

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
                readOnly={isChatReadOnly}
              />
            </Box>
          )}
        </Box>
      )}
    </Box>
  );
}

export default DealPage;
