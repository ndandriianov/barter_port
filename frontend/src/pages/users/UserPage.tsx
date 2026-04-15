import { useMemo, useState } from "react";
import { Link as RouterLink, Navigate, useNavigate, useParams } from "react-router-dom";
import {
  Alert,
  Avatar,
  Box,
  Button,
  Card,
  CardContent,
  CircularProgress,
  Divider,
  Snackbar,
  Stack,
  Typography,
} from "@mui/material";
import PersonOutlineIcon from "@mui/icons-material/PersonOutline";
import usersApi from "@/features/users/api/usersApi.ts";

function UserPage() {
  const { userId } = useParams<{ userId: string }>();
  const navigate = useNavigate();
  const { data: currentUser, isLoading: isCurrentUserLoading } = usersApi.useGetCurrentUserQuery();
  const { data: user, isLoading: isUserLoading, error } = usersApi.useGetUserByIdQuery(userId ?? "", {
    skip: !userId,
  });
  const { data: subscriptions, isLoading: isSubscriptionsLoading } = usersApi.useGetSubscriptionsQuery();
  const [subscribeToUser, { isLoading: isSubscribing }] = usersApi.useSubscribeToUserMutation();
  const [unsubscribeFromUser, { isLoading: isUnsubscribing }] = usersApi.useUnsubscribeFromUserMutation();
  const [snackbarMessage, setSnackbarMessage] = useState<string | null>(null);

  const isSubscribed = useMemo(() => {
    if (!subscriptions || !userId) {
      return false;
    }
    return subscriptions.some((sub) => sub.id === userId);
  }, [subscriptions, userId]);

  const handleSubscribe = async () => {
    if (!userId) {
      return;
    }
    try {
      if (isSubscribed) {
        await unsubscribeFromUser({ targetUserId: userId }).unwrap();
        setSnackbarMessage("Вы успешно отписались от пользователя");
      } else {
        await subscribeToUser({ targetUserId: userId }).unwrap();
        setSnackbarMessage("Вы успешно подписались на пользователя");
      }
    } catch {
      setSnackbarMessage(isSubscribed 
        ? "Не удалось отписаться от пользователя" 
        : "Не удалось подписаться на пользователя");
    }
  };

  if (!userId) {
    return <Alert severity="warning">Пользователь не найден</Alert>;
  }

  if (isCurrentUserLoading || isUserLoading || isSubscriptionsLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (currentUser?.id === userId) {
    return <Navigate to="/profile" replace />;
  }

  if (error || !user) {
    return <Alert severity="warning">Пользователь не найден</Alert>;
  }

  const displayName = user.name?.trim() || "Имя не указано";
  const bio = user.bio?.trim();
  const avatarUrl = user.avatarUrl?.trim() || "";

  return (
    <Box maxWidth={560} mx="auto">
      <Button
        size="small"
        variant="text"
        onClick={() => window.history.length > 1 ? navigate(-1) : navigate("/offers")}
        sx={{ mb: 2 }}
      >
        ← Назад
      </Button>

      <Typography variant="h4" fontWeight={700} mb={3}>
        Профиль пользователя
      </Typography>

      <Card variant="outlined">
        <CardContent>
          <Box display="flex" alignItems="center" gap={2} mb={3}>
            <Avatar
              src={avatarUrl || undefined}
              alt={displayName}
              sx={{ width: 72, height: 72, bgcolor: "action.selected" }}
            >
              {!avatarUrl && <PersonOutlineIcon fontSize="large" color="action" />}
            </Avatar>
            <Box>
              <Typography variant="h5" fontWeight={700}>
                {displayName}
              </Typography>
              <Typography variant="caption" color="text.secondary">
                ID
              </Typography>
              <Typography variant="body2" fontFamily="monospace" fontWeight={500}>
                {user.id}
              </Typography>
            </Box>
          </Box>

          <Stack spacing={1.5} mb={3}>
            <Box>
              <Typography variant="caption" color="text.secondary">
                О пользователе
              </Typography>
              <Typography variant="body2" sx={{ whiteSpace: "pre-wrap" }}>
                {bio || "Пользователь пока ничего о себе не рассказал."}
              </Typography>
            </Box>
          </Stack>

          <Divider sx={{ my: 3 }} />

          <Box display="flex" gap={2} flexWrap="wrap">
            <Button
              variant={isSubscribed ? "outlined" : "contained"}
              onClick={handleSubscribe}
              disabled={isSubscribing || isUnsubscribing}
            >
              {isSubscribed ? "Отписаться" : "Подписаться"}
            </Button>
            <Button component={RouterLink} to={`/users/${user.id}/reviews`} variant="outlined">
              Отзывы о пользователе
            </Button>
          </Box>
        </CardContent>
      </Card>

      <Snackbar
        open={snackbarMessage !== null}
        autoHideDuration={3000}
        onClose={() => setSnackbarMessage(null)}
        message={snackbarMessage}
      />
    </Box>
  );
}

export default UserPage;
