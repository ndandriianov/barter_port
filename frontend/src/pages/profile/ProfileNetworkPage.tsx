import { Alert, Avatar, Box, Button, ButtonGroup, Card, CardContent, List, ListItem, ListItemAvatar, ListItemText, Stack, Typography } from "@mui/material";
import PersonOutlineIcon from "@mui/icons-material/PersonOutline";
import { Link as RouterLink } from "react-router-dom";
import usersApi from "@/features/users/api/usersApi.ts";
import { appRoutes } from "@/shared/config/appRoutes.ts";
import type { User } from "@/features/users/model/types.ts";
import ProfileSectionShell from "@/widgets/profile/ProfileSectionShell.tsx";

interface ProfileNetworkPageProps {
  mode: "subscriptions" | "subscribers" | "hidden";
}

function ProfileNetworkPage({ mode }: ProfileNetworkPageProps) {
  const subscriptionsQuery = usersApi.useGetSubscriptionsQuery(undefined, {
    skip: mode !== "subscriptions",
  });
  const subscribersQuery = usersApi.useGetSubscribersQuery(undefined, {
    skip: mode !== "subscribers",
  });
  const hiddenUsersQuery = usersApi.useGetHiddenUsersQuery(undefined, {
    skip: mode !== "hidden",
  });

  const activeQuery = mode === "subscriptions"
    ? subscriptionsQuery
    : mode === "subscribers"
      ? subscribersQuery
      : hiddenUsersQuery;
  const {
    data,
    isLoading,
    error,
    refetch,
    isFetching,
  } = activeQuery;

  const title = mode === "subscriptions" ? "Подписки" : mode === "subscribers" ? "Подписчики" : "Черный список";
  const description = mode === "subscriptions"
    ? "Люди, на которых вы подписаны"
    : mode === "subscribers"
      ? "Люди, которые подписались на вас. Взаимная подписка является условием создания нового личного чата."
      : "Авторы, чьи объявления скрыты из общего каталога для вашего аккаунта.";

  const renderUserListItem = (user: User) => (
    <ListItem
      key={user.id}
      component={RouterLink}
      to={`/users/${user.id}`}
      sx={{ textDecoration: "none", color: "inherit" }}
    >
      <ListItemAvatar>
        <Avatar src={user.avatarUrl?.trim() || undefined} sx={{ width: 44, height: 44 }}>
          {!user.avatarUrl?.trim() && <PersonOutlineIcon fontSize="small" />}
        </Avatar>
      </ListItemAvatar>
      <ListItemText
        primary={user.name?.trim() || "Имя не указано"}
      />
    </ListItem>
  );

  return (
    <ProfileSectionShell
      title={title}
      description={description}
      actions={
        <Button variant="outlined" onClick={() => refetch()} disabled={isFetching}>
          Обновить
        </Button>
      }
    >
      <Stack spacing={3}>
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
            to={appRoutes.profile.networkSubscriptions}
            variant={mode === "subscriptions" ? "contained" : "text"}
          >
            Подписки
          </Button>
          <Button
            component={RouterLink}
            to={appRoutes.profile.networkSubscribers}
            variant={mode === "subscribers" ? "contained" : "text"}
          >
            Подписчики
          </Button>
          <Button
            component={RouterLink}
            to={appRoutes.profile.networkHidden}
            variant={mode === "hidden" ? "contained" : "text"}
          >
            Черный список
          </Button>
        </ButtonGroup>

        {mode === "hidden" ? (
          <Alert severity="info">
            Пользователи из черного списка не исчезают из приложения полностью, но их объявления
            скрываются из общего списка. Чтобы подписаться на такого пользователя, сначала уберите его
            из черного списка.
          </Alert>
        ) : (
          <Alert severity="info">
            Новый личный чат можно создать только при взаимной подписке. Если подписка разорвана,
            существующий чат остаётся рабочим, но создать новый уже нельзя.
          </Alert>
        )}

        <Card variant="outlined">
          <CardContent>
            {isLoading ? (
              <Typography color="text.secondary">Загрузка списка...</Typography>
            ) : error ? (
              <Alert severity="error">Не удалось загрузить этот список.</Alert>
            ) : !data || data.length === 0 ? (
              <Alert severity="info">
                {mode === "subscriptions"
                  ? "Вы пока ни на кого не подписаны."
                  : mode === "subscribers"
                    ? "У вас пока нет подписчиков."
                    : "Черный список пока пуст."}
              </Alert>
            ) : (
              <Box>
                <Typography variant="h6" fontWeight={800} mb={2}>
                  {title}: {data.length}
                </Typography>
                <List>
                  {data.map((user) => renderUserListItem(user))}
                </List>
              </Box>
            )}
          </CardContent>
        </Card>
      </Stack>
    </ProfileSectionShell>
  );
}

export default ProfileNetworkPage;
