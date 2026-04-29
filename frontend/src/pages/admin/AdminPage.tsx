import {type ReactNode, useMemo, useState} from "react";
import {Link as RouterLink} from "react-router-dom";
import {skipToken} from "@reduxjs/toolkit/query";
import type {FetchBaseQueryError} from "@reduxjs/toolkit/query";
import {
  Alert,
  Avatar,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Grid,
  Stack,
  TextField,
  Typography,
} from "@mui/material";
import AdminPanelSettingsOutlinedIcon from "@mui/icons-material/AdminPanelSettingsOutlined";
import ForumOutlinedIcon from "@mui/icons-material/ForumOutlined";
import Inventory2OutlinedIcon from "@mui/icons-material/Inventory2Outlined";
import ManageSearchOutlinedIcon from "@mui/icons-material/ManageSearchOutlined";
import PeopleAltOutlinedIcon from "@mui/icons-material/PeopleAltOutlined";
import SellOutlinedIcon from "@mui/icons-material/SellOutlined";
import ShieldOutlinedIcon from "@mui/icons-material/ShieldOutlined";
import TagOutlinedIcon from "@mui/icons-material/TagOutlined";
import VerifiedOutlinedIcon from "@mui/icons-material/VerifiedOutlined";
import VisibilityOutlinedIcon from "@mui/icons-material/VisibilityOutlined";
import authApi from "@/features/auth/api/authApi.ts";
import chatsApi from "@/features/chats/api/chatsApi.ts";
import dealsApi from "@/features/deals/api/dealsApi.ts";
import offersApi from "@/features/offers/api/offersApi.ts";
import usersApi from "@/features/users/api/usersApi.ts";
import {appRoutes} from "@/shared/config/appRoutes.ts";

const quickActions = [
  { label: "Жалобы", to: appRoutes.admin.offerReports },
  { label: "Провалы сделок", to: appRoutes.admin.failures },
  { label: "Объявления", to: appRoutes.market.catalog },
] as const;

const numberFormatter = new Intl.NumberFormat("ru-RU");
const percentFormatter = new Intl.NumberFormat("ru-RU", {
  style: "percent",
  maximumFractionDigits: 1,
});
const dateTimeFormatter = new Intl.DateTimeFormat("ru-RU", {
  dateStyle: "medium",
  timeStyle: "short",
});
const ratingFormatter = new Intl.NumberFormat("ru-RU", {
  minimumFractionDigits: 1,
  maximumFractionDigits: 1,
});

interface MetricCardProps {
  title: string;
  value: string;
  caption: string;
  icon: ReactNode;
  isLoading?: boolean;
}

function MetricCard({ title, value, caption, icon, isLoading = false }: MetricCardProps) {
  return (
    <Card variant="outlined" sx={{ height: "100%", borderRadius: 2 }}>
      <CardContent>
        <Stack spacing={1.5}>
          <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2}>
            <Typography variant="body2" color="text.secondary">
              {title}
            </Typography>
            <Typography color="primary.main">{icon}</Typography>
          </Box>
          {isLoading ? (
            <CircularProgress size={24} />
          ) : (
            <Typography variant="h4" fontWeight={800}>
              {value}
            </Typography>
          )}
          <Typography variant="body2" color="text.secondary">
            {caption}
          </Typography>
        </Stack>
      </CardContent>
    </Card>
  );
}

interface StatLineProps {
  label: string;
  value: string;
}

function StatLine({ label, value }: StatLineProps) {
  return (
    <Box display="flex" justifyContent="space-between" gap={2}>
      <Typography variant="body2" color="text.secondary">
        {label}
      </Typography>
      <Typography variant="body2" fontWeight={700} textAlign="right">
        {value}
      </Typography>
    </Box>
  );
}

function formatNumber(value: number | bigint) {
  return numberFormatter.format(value);
}

function formatPercent(value: number) {
  return percentFormatter.format(value);
}

function formatRating(value: number | null | undefined) {
  if (value === null || value === undefined) {
    return "Нет данных";
  }

  return ratingFormatter.format(value);
}

function formatDateTime(value: string) {
  return dateTimeFormatter.format(new Date(value));
}

function formatName(name: string | undefined, fallback: string) {
  const trimmed = name?.trim();
  return trimmed && trimmed.length > 0 ? trimmed : fallback;
}

function getErrorMessage(error: unknown, fallback: string) {
  if (!error || typeof error !== "object") {
    return fallback;
  }

  if ("status" in error) {
    const fetchError = error as FetchBaseQueryError & { data?: unknown };
    if (fetchError.data && typeof fetchError.data === "object" && "message" in fetchError.data) {
      const message = (fetchError.data as { message?: unknown }).message;
      if (typeof message === "string" && message.trim().length > 0) {
        return message;
      }
    }

    if (fetchError.status === 404) {
      return "Пользователь не найден.";
    }
  }

  return fallback;
}

function AdminPage() {
  const { data: currentUser, isLoading: isCurrentUserLoading } = usersApi.useGetCurrentUserQuery();
  const isAdmin = currentUser?.isAdmin === true;

  const { data: authPlatform, isLoading: isAuthPlatformLoading, error: authPlatformError } =
    authApi.useGetAdminPlatformStatisticsQuery(isAdmin ? undefined : skipToken);
  const { data: usersPlatform, isLoading: isUsersPlatformLoading, error: usersPlatformError } =
    usersApi.useGetAdminPlatformStatisticsQuery(isAdmin ? undefined : skipToken);
  const { data: chatsPlatform, isLoading: isChatsPlatformLoading, error: chatsPlatformError } =
    chatsApi.useGetAdminPlatformStatisticsQuery(isAdmin ? undefined : skipToken);
  const { data: dealsPlatform, isLoading: isDealsPlatformLoading, error: dealsPlatformError } =
    dealsApi.useGetAdminPlatformStatisticsQuery(isAdmin ? undefined : skipToken);
  const { data: adminUsers = [], isLoading: isAdminUsersLoading, error: adminUsersError } =
    usersApi.useListAdminUsersQuery(isAdmin ? undefined : skipToken);
  const { data: tags = [], isLoading: isTagsLoading } =
    offersApi.useListTagsQuery(isAdmin ? undefined : skipToken);
  const [deleteAdminTag, { isLoading: isDeletingTag }] = offersApi.useDeleteAdminTagMutation();

  const [userSearch, setUserSearch] = useState("");
  const [selectedUserId, setSelectedUserId] = useState<string | null>(null);
  const normalizedUserSearch = userSearch.trim().toLowerCase();
  const resolvedUserId = selectedUserId;

  const {
    data: authUserStats,
    isLoading: isAuthUserStatsLoading,
    error: authUserStatsError,
  } = authApi.useGetAdminUserStatisticsQuery(resolvedUserId ?? skipToken);
  const {
    data: usersUserStats,
    isLoading: isUsersUserStatsLoading,
    error: usersUserStatsError,
  } = usersApi.useGetAdminUserStatisticsQuery(resolvedUserId ?? skipToken);
  const {
    data: dealsUserStats,
    isLoading: isDealsUserStatsLoading,
    error: dealsUserStatsError,
  } = dealsApi.useGetAdminUserStatisticsQuery(resolvedUserId ?? skipToken);

  const filteredUsers = useMemo(() => {
    if (!normalizedUserSearch) {
      return adminUsers;
    }

    return adminUsers.filter((user) =>
      [user.id, user.name, user.phoneNumber, user.bio].some(
        (value) => value?.toLowerCase().includes(normalizedUserSearch),
      ),
    );
  }, [adminUsers, normalizedUserSearch]);

  const inspectedUser = adminUsers.find((user) => user.id === resolvedUserId) ?? null;
  const isPlatformLoading =
    isAuthPlatformLoading ||
    isUsersPlatformLoading ||
    isChatsPlatformLoading ||
    isDealsPlatformLoading;
  const isUserDetailsLoading =
    resolvedUserId !== null && (isAuthUserStatsLoading || isUsersUserStatsLoading || isDealsUserStatsLoading);

  const platformErrorMessage = useMemo(() => {
    return (
      getErrorMessage(authPlatformError, "") ||
      getErrorMessage(usersPlatformError, "") ||
      getErrorMessage(chatsPlatformError, "") ||
      getErrorMessage(dealsPlatformError, "")
    );
  }, [authPlatformError, chatsPlatformError, dealsPlatformError, usersPlatformError]);

  const adminUsersErrorMessage = useMemo(() => {
    return getErrorMessage(adminUsersError, "");
  }, [adminUsersError]);

  const userErrorMessage = useMemo(() => {
    return (
      getErrorMessage(authUserStatsError, "") ||
      getErrorMessage(usersUserStatsError, "") ||
      getErrorMessage(dealsUserStatsError, "")
    );
  }, [authUserStatsError, dealsUserStatsError, usersUserStatsError]);

  if (isCurrentUserLoading) {
    return (
      <Box display="flex" justifyContent="center" py={8}>
        <CircularProgress />
      </Box>
    );
  }

  if (!currentUser) {
    return <Alert severity="warning">Не удалось получить данные текущего пользователя.</Alert>;
  }

  if (!currentUser.isAdmin) {
    return (
      <Stack spacing={3}>
        <Alert severity="error">Доступ к разделу статистики есть только у администратора.</Alert>
        <Button component={RouterLink} to={appRoutes.profile.home} variant="contained" sx={{ alignSelf: "flex-start" }}>
          Вернуться в профиль
        </Button>
      </Stack>
    );
  }

  return (
    <Stack spacing={4}>
      <Box
        sx={{
          px: { xs: 3, md: 4 },
          py: { xs: 3, md: 4 },
          borderRadius: 2,
          color: "common.white",
          background:
            "linear-gradient(135deg, rgba(19,47,76,1) 0%, rgba(24,86,122,1) 58%, rgba(24,125,100,1) 100%)",
        }}
      >
        <Stack spacing={2.5}>
          <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2} flexWrap="wrap">
            <Box>
              <Box display="flex" alignItems="center" gap={1} mb={1}>
                <AdminPanelSettingsOutlinedIcon fontSize="small" />
                <Typography variant="overline" sx={{ opacity: 0.92 }}>
                  Admin Statistics
                </Typography>
              </Box>
              <Typography variant="h4" fontWeight={800} mb={1}>
                Статистика по платформе и отдельно по каждому пользователю
              </Typography>
            </Box>
            <Chip
              label={`admin: ${formatName(currentUser.name, currentUser.email)}`}
              sx={{
                bgcolor: "rgba(255,255,255,0.12)",
                color: "common.white",
                borderRadius: 2,
                fontWeight: 700,
              }}
            />
          </Box>

          <Box display="flex" gap={1.25} flexWrap="wrap">
            {quickActions.map((action) => (
              <Button
                key={action.to}
                component={RouterLink}
                to={action.to}
                variant="contained"
                sx={{
                  bgcolor: "rgba(255,255,255,0.12)",
                  "&:hover": { bgcolor: "rgba(255,255,255,0.2)" },
                }}
              >
                {action.label}
              </Button>
            ))}
          </Box>
        </Stack>
      </Box>

      {platformErrorMessage ? (
        <Alert severity="warning">
          Не удалось загрузить часть платформенной статистики. {platformErrorMessage}
        </Alert>
      ) : null}

      <Grid container spacing={2}>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <MetricCard
            title="Зарегистрированные аккаунты"
            value={formatNumber(authPlatform?.users.totalRegistered ?? 0)}
            caption="Все аккаунты из auth."
            icon={<PeopleAltOutlinedIcon />}
            isLoading={isAuthPlatformLoading}
          />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <MetricCard
            title="Подтверждённые email"
            value={formatNumber(authPlatform?.users.verifiedEmails ?? 0)}
            caption="Почты с завершённой верификацией."
            icon={<VerifiedOutlinedIcon />}
            isLoading={isAuthPlatformLoading}
          />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <MetricCard
            title="Чаты"
            value={formatNumber(chatsPlatform?.chats.total ?? 0)}
            caption="Общее число диалогов на платформе."
            icon={<ForumOutlinedIcon />}
            isLoading={isChatsPlatformLoading}
          />
        </Grid>
        <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
          <MetricCard
            title="Сделки"
            value={formatNumber(dealsPlatform?.deals.total ?? 0)}
            caption="Все сделки, включая активные и терминальные."
            icon={<Inventory2OutlinedIcon />}
            isLoading={isDealsPlatformLoading}
          />
        </Grid>
      </Grid>

      <Grid container spacing={2}>
        <Grid size={{ xs: 12, lg: 4 }}>
          <Card variant="outlined" sx={{ height: "100%", borderRadius: 2 }}>
            <CardContent>
              <Stack spacing={2}>
                <Box display="flex" alignItems="center" gap={1}>
                  <ShieldOutlinedIcon color="primary" />
                  <Typography variant="h6" fontWeight={800}>
                    Auth
                  </Typography>
                </Box>
                <StatLine
                  label="Всего пользователей"
                  value={formatNumber(authPlatform?.users.totalRegistered ?? 0)}
                />
                <StatLine
                  label="Подтверждённые email"
                  value={formatNumber(authPlatform?.users.verifiedEmails ?? 0)}
                />
                <StatLine
                  label="Доля подтверждений"
                  value={
                    authPlatform && authPlatform.users.totalRegistered > 0
                      ? formatPercent(authPlatform.users.verifiedEmails / authPlatform.users.totalRegistered)
                      : "0%"
                  }
                />
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        <Grid size={{ xs: 12, lg: 4 }}>
          <Card variant="outlined" sx={{ height: "100%", borderRadius: 2 }}>
            <CardContent>
              <Stack spacing={2}>
                <Box display="flex" alignItems="center" gap={1}>
                  <PeopleAltOutlinedIcon color="primary" />
                  <Typography variant="h6" fontWeight={800}>
                    Users
                  </Typography>
                </Box>
                <StatLine label="Средняя репутация" value={formatRating(usersPlatform?.reputation.average)} />
                <StatLine label="Медианная репутация" value={formatRating(usersPlatform?.reputation.median)} />
                <Box>
                  <Typography variant="body2" color="text.secondary" mb={1}>
                    Топ по репутации
                  </Typography>
                  <Stack spacing={0.75}>
                    {(usersPlatform?.reputation.topUsers ?? []).slice(0, 5).map((user) => (
                      <StatLine
                        key={user.userId}
                        label={formatName(user.name, user.userId)}
                        value={formatNumber(user.reputationPoints)}
                      />
                    ))}
                  </Stack>
                </Box>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        <Grid size={{ xs: 12, lg: 4 }}>
          <Card variant="outlined" sx={{ height: "100%", borderRadius: 2 }}>
            <CardContent>
              <Stack spacing={2}>
                <Box display="flex" alignItems="center" gap={1}>
                  <ForumOutlinedIcon color="primary" />
                  <Typography variant="h6" fontWeight={800}>
                    Chats
                  </Typography>
                </Box>
                <StatLine label="Всего чатов" value={formatNumber(chatsPlatform?.chats.total ?? 0)} />
                <Typography variant="body2" color="text.secondary">
                  Метрика идёт из отдельного chats сервиса и не зависит от выборок списка сообщений на клиенте.
                </Typography>
              </Stack>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Grid container spacing={2}>
        <Grid size={{ xs: 12, xl: 6 }}>
          <Card variant="outlined" sx={{ height: "100%", borderRadius: 2 }}>
            <CardContent>
              <Stack spacing={2}>
                <Box display="flex" alignItems="center" gap={1}>
                  <SellOutlinedIcon color="primary" />
                  <Typography variant="h6" fontWeight={800}>
                    Объявления
                  </Typography>
                </Box>
                <StatLine label="Всего" value={formatNumber(dealsPlatform?.offers.total ?? 0)} />
                <StatLine label="Черновики сделок" value={formatNumber(dealsPlatform?.offers.drafts ?? 0)} />
                <StatLine label="Просмотры" value={formatNumber(dealsPlatform?.offers.totalViews ?? 0)} />
                <StatLine
                  label="Среднее на пользователя"
                  value={formatRating(dealsPlatform?.offers.averagePerUser)}
                />
                <StatLine label="Средний рейтинг" value={formatRating(dealsPlatform?.offers.averageRating)} />
                <StatLine
                  label="Скрыто модератором"
                  value={formatNumber(dealsPlatform?.offers.hidden.moderated ?? 0)}
                />
                <StatLine
                  label="Скрыто автором"
                  value={formatNumber(dealsPlatform?.offers.hidden.hiddenByAuthor ?? 0)}
                />
                <StatLine
                  label="good / service"
                  value={`${formatNumber(dealsPlatform?.offers.byType.good ?? 0)} / ${formatNumber(dealsPlatform?.offers.byType.service ?? 0)}`}
                />
                <StatLine
                  label="give / take"
                  value={`${formatNumber(dealsPlatform?.offers.byAction.give ?? 0)} / ${formatNumber(dealsPlatform?.offers.byAction.take ?? 0)}`}
                />
                <Box>
                  <Typography variant="body2" color="text.secondary" mb={1}>
                    Топ тегов
                  </Typography>
                  <Box display="flex" flexWrap="wrap" gap={1}>
                    {(dealsPlatform?.offers.topTags ?? []).slice(0, 6).map((tag) => (
                      <Chip
                        key={tag.tag}
                        label={`${tag.tag} · ${formatNumber(tag.offersCount)}`}
                        size="small"
                        variant="outlined"
                      />
                    ))}
                  </Box>
                </Box>
              </Stack>
            </CardContent>
          </Card>
        </Grid>

        <Grid size={{ xs: 12, xl: 6 }}>
          <Card variant="outlined" sx={{ height: "100%", borderRadius: 2 }}>
            <CardContent>
              <Stack spacing={2}>
                <Box display="flex" alignItems="center" gap={1}>
                  <Inventory2OutlinedIcon color="primary" />
                  <Typography variant="h6" fontWeight={800}>
                    Сделки, жалобы и отзывы
                  </Typography>
                </Box>
                <StatLine label="Сделок всего" value={formatNumber(dealsPlatform?.deals.total ?? 0)} />
                <StatLine
                  label="Успешная конверсия"
                  value={formatPercent(dealsPlatform?.deals.successfulConversionRate ?? 0)}
                />
                <StatLine
                  label="Среднее число участников"
                  value={formatRating(dealsPlatform?.deals.averageParticipants)}
                />
                <StatLine
                  label="Доля multi-party"
                  value={formatPercent(dealsPlatform?.deals.multiPartyShare ?? 0)}
                />
                <Box>
                  <Typography variant="body2" color="text.secondary" mb={1}>
                    Статусы сделок
                  </Typography>
                  <Box display="flex" flexWrap="wrap" gap={1}>
                    <Chip
                      size="small"
                      variant="outlined"
                      label={`LFP ${formatNumber(dealsPlatform?.deals.byStatus.lookingForParticipants ?? 0)}`}
                    />
                    <Chip
                      size="small"
                      variant="outlined"
                      label={`Discussion ${formatNumber(dealsPlatform?.deals.byStatus.discussion ?? 0)}`}
                    />
                    <Chip
                      size="small"
                      variant="outlined"
                      label={`Confirmed ${formatNumber(dealsPlatform?.deals.byStatus.confirmed ?? 0)}`}
                    />
                    <Chip
                      size="small"
                      variant="outlined"
                      label={`Completed ${formatNumber(dealsPlatform?.deals.byStatus.completed ?? 0)}`}
                    />
                    <Chip
                      size="small"
                      variant="outlined"
                      label={`Failed ${formatNumber(dealsPlatform?.deals.byStatus.failed ?? 0)}`}
                    />
                    <Chip
                      size="small"
                      variant="outlined"
                      label={`Cancelled ${formatNumber(dealsPlatform?.deals.byStatus.cancelled ?? 0)}`}
                    />
                  </Box>
                </Box>
                <StatLine label="Жалоб всего" value={formatNumber(dealsPlatform?.reports.total ?? 0)} />
                <StatLine
                  label="Жалоб в ожидании"
                  value={formatNumber(dealsPlatform?.reports.pending ?? 0)}
                />
                <StatLine
                  label="Заблокированные объявления"
                  value={formatNumber(dealsPlatform?.reports.blockedOffers ?? 0)}
                />
                <StatLine
                  label="Решения по failure"
                  value={formatNumber(dealsPlatform?.reports.adminFailureResolutions ?? 0)}
                />
                <StatLine label="Отзывов всего" value={formatNumber(dealsPlatform?.reviews.total ?? 0)} />
                <StatLine
                  label="Средний рейтинг отзывов"
                  value={formatRating(dealsPlatform?.reviews.averageRating)}
                />
              </Stack>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Box
        sx={{
          border: "1px solid",
          borderColor: "divider",
          borderRadius: 2,
          px: { xs: 2, md: 3 },
          py: { xs: 2, md: 3 },
        }}
      >
        <Stack spacing={3}>
          <Box display="flex" alignItems="center" gap={1}>
            <ManageSearchOutlinedIcon color="primary" />
            <Typography variant="h6" fontWeight={800}>
              Проверка пользователя
            </Typography>
          </Box>
          <Typography variant="body2" color="text.secondary">
            Слева список пользователей из отдельного endpoint `users` service. После выбора подтягиваются детали из
            `auth`, `users` и `deals`.
          </Typography>

          {adminUsersErrorMessage ? (
            <Alert severity="warning">
              Не удалось загрузить список пользователей. {adminUsersErrorMessage}
            </Alert>
          ) : null}

          <Grid container spacing={2}>
            <Grid size={{ xs: 12, lg: 4 }}>
              <Card variant="outlined" sx={{ height: "100%", borderRadius: 2 }}>
                <CardContent>
                  <Stack spacing={2}>
                    <TextField
                      fullWidth
                      label="Поиск пользователя"
                      value={userSearch}
                      onChange={(event) => setUserSearch(event.target.value)}
                      helperText="По имени, телефону, bio или UUID."
                    />

                    {isAdminUsersLoading ? (
                      <Box display="flex" justifyContent="center" py={4}>
                        <CircularProgress />
                      </Box>
                    ) : filteredUsers.length === 0 ? (
                      <Alert severity="info">
                        {adminUsers.length === 0 ? "Список пользователей пуст." : "Ничего не найдено по текущему фильтру."}
                      </Alert>
                    ) : (
                      <Stack spacing={1} sx={{ maxHeight: 540, overflowY: "auto", pr: 0.5 }}>
                        {filteredUsers.map((user) => {
                          const isSelected = user.id === resolvedUserId;

                          return (
                            <Button
                              key={user.id}
                              fullWidth
                              variant="outlined"
                              onClick={() => setSelectedUserId(user.id)}
                              sx={{
                                justifyContent: "flex-start",
                                p: 1.5,
                                textTransform: "none",
                                borderRadius: 2,
                                borderColor: isSelected ? "primary.main" : "divider",
                                bgcolor: isSelected ? "action.selected" : "background.paper",
                              }}
                            >
                              <Box display="flex" alignItems="flex-start" gap={1.5} width="100%">
                                <Avatar src={user.avatarUrl} sx={{ width: 40, height: 40 }}>
                                  {formatName(user.name, user.id).slice(0, 1).toUpperCase()}
                                </Avatar>
                                <Box minWidth={0} flexGrow={1} textAlign="left">
                                  <Box display="flex" alignItems="center" justifyContent="space-between" gap={1} mb={0.5}>
                                    <Typography variant="body2" fontWeight={700} noWrap>
                                      {formatName(user.name, user.id)}
                                    </Typography>
                                    <Chip
                                      size="small"
                                      label={formatNumber(user.reputationPoints)}
                                      color={isSelected ? "primary" : "default"}
                                    />
                                  </Box>
                                  <Typography variant="caption" color="text.secondary" sx={{ wordBreak: "break-all" }}>
                                    {user.id}
                                  </Typography>
                                  {user.phoneNumber ? (
                                    <Typography variant="caption" color="text.secondary" display="block">
                                      {user.phoneNumber}
                                    </Typography>
                                  ) : null}
                                </Box>
                              </Box>
                            </Button>
                          );
                        })}
                      </Stack>
                    )}
                  </Stack>
                </CardContent>
              </Card>
            </Grid>

            <Grid size={{ xs: 12, lg: 8 }}>
              {resolvedUserId === null ? (
                <Card variant="outlined" sx={{ borderRadius: 2 }}>
                  <CardContent>
                    <Alert severity="info">Выберите пользователя из списка слева, чтобы открыть его данные и статистику.</Alert>
                  </CardContent>
                </Card>
              ) : (
                <Stack spacing={2.5}>
                  {userErrorMessage ? <Alert severity="warning">{userErrorMessage}</Alert> : null}

                  {isUserDetailsLoading ? (
                    <Box display="flex" justifyContent="center" py={6}>
                      <CircularProgress />
                    </Box>
                  ) : authUserStats && usersUserStats && dealsUserStats ? (
                    <>
                      <Grid container spacing={2}>
                        <Grid size={{ xs: 12, lg: 4 }}>
                          <Card variant="outlined" sx={{ height: "100%", borderRadius: 2 }}>
                            <CardContent>
                              <Stack spacing={1.5}>
                                <Typography variant="body2" color="text.secondary">
                                  Профиль
                                </Typography>
                                <Box display="flex" alignItems="flex-start" gap={1.5}>
                                  <Avatar src={inspectedUser?.avatarUrl} sx={{ width: 52, height: 52 }}>
                                    {formatName(inspectedUser?.name, authUserStats.userId).slice(0, 1).toUpperCase()}
                                  </Avatar>
                                  <Box minWidth={0}>
                                    <Typography variant="h6" fontWeight={800}>
                                      {formatName(inspectedUser?.name, authUserStats.userId)}
                                    </Typography>
                                    <Typography variant="body2" color="text.secondary" sx={{ wordBreak: "break-all" }}>
                                      {authUserStats.userId}
                                    </Typography>
                                  </Box>
                                </Box>
                                <StatLine label="Регистрация" value={formatDateTime(authUserStats.registeredAt)} />
                                <StatLine label="Телефон" value={inspectedUser?.phoneNumber ?? "Не указан"} />
                                <StatLine
                                  label="Email"
                                  value={authUserStats.emailVerified ? "Подтверждён" : "Не подтверждён"}
                                />
                                <Typography variant="body2" color="text.secondary">
                                  {inspectedUser?.bio?.trim() || "Био не заполнено."}
                                </Typography>
                              </Stack>
                            </CardContent>
                          </Card>
                        </Grid>

                        <Grid size={{ xs: 12, lg: 4 }}>
                          <Card variant="outlined" sx={{ height: "100%", borderRadius: 2 }}>
                            <CardContent>
                              <Stack spacing={1.5}>
                                <Typography variant="body2" color="text.secondary">
                                  Репутация и social
                                </Typography>
                                <StatLine
                                  label="Текущая репутация"
                                  value={formatNumber(usersUserStats.reputation.currentPoints)}
                                />
                                <StatLine
                                  label="Подписчики"
                                  value={formatNumber(usersUserStats.social.followersCount)}
                                />
                                <StatLine
                                  label="Подписки"
                                  value={formatNumber(usersUserStats.social.subscriptionsCount)}
                                />
                                <StatLine
                                  label="События в истории"
                                  value={formatNumber(usersUserStats.reputation.history.length)}
                                />
                              </Stack>
                            </CardContent>
                          </Card>
                        </Grid>

                        <Grid size={{ xs: 12, lg: 4 }}>
                          <Card variant="outlined" sx={{ height: "100%", borderRadius: 2 }}>
                            <CardContent>
                              <Stack spacing={1.5}>
                                <Typography variant="body2" color="text.secondary">
                                  Активность в deals
                                </Typography>
                                <StatLine
                                  label="Опубликованные объявления"
                                  value={formatNumber(dealsUserStats.offers.published)}
                                />
                                <StatLine
                                  label="Просмотры объявлений"
                                  value={formatNumber(dealsUserStats.offers.totalViews)}
                                />
                                <StatLine label="Активные сделки" value={formatNumber(dealsUserStats.deals.active)} />
                                <StatLine
                                  label="Завершённые сделки"
                                  value={formatNumber(dealsUserStats.deals.completed)}
                                />
                              </Stack>
                            </CardContent>
                          </Card>
                        </Grid>
                      </Grid>

                      <Grid container spacing={2}>
                        <Grid size={{ xs: 12, xl: 4 }}>
                          <Card variant="outlined" sx={{ height: "100%", borderRadius: 2 }}>
                            <CardContent>
                              <Stack spacing={1.5}>
                                <Typography variant="h6" fontWeight={800}>
                                  События репутации
                                </Typography>
                                {usersUserStats.reputation.history.length === 0 ? (
                                  <Typography variant="body2" color="text.secondary">
                                    История пока пустая.
                                  </Typography>
                                ) : (
                                  usersUserStats.reputation.history.slice(0, 6).map((event) => (
                                    <Box key={event.id} sx={{ py: 0.5 }}>
                                      <Box display="flex" justifyContent="space-between" gap={2}>
                                        <Typography variant="body2" fontWeight={700}>
                                          {event.delta > 0 ? `+${event.delta}` : event.delta}
                                        </Typography>
                                        <Typography variant="caption" color="text.secondary">
                                          {formatDateTime(event.createdAt)}
                                        </Typography>
                                      </Box>
                                      <Typography variant="body2" color="text.secondary">
                                        {event.sourceType}
                                      </Typography>
                                    </Box>
                                  ))
                                )}
                              </Stack>
                            </CardContent>
                          </Card>
                        </Grid>

                        <Grid size={{ xs: 12, xl: 4 }}>
                          <Card variant="outlined" sx={{ height: "100%", borderRadius: 2 }}>
                            <CardContent>
                              <Stack spacing={1.5}>
                                <Typography variant="h6" fontWeight={800}>
                                  Сделки и отзывы
                                </Typography>
                                <StatLine label="Failed всего" value={formatNumber(dealsUserStats.deals.failed.total)} />
                                <StatLine
                                  label="Failed по вине пользователя"
                                  value={formatNumber(dealsUserStats.deals.failed.responsible)}
                                />
                                <StatLine
                                  label="Failed как пострадавший"
                                  value={formatNumber(dealsUserStats.deals.failed.affected)}
                                />
                                <StatLine label="Cancelled" value={formatNumber(dealsUserStats.deals.cancelled)} />
                                <StatLine label="Получено отзывов" value={formatNumber(dealsUserStats.reviews.received)} />
                                <StatLine
                                  label="Средняя оценка"
                                  value={formatRating(dealsUserStats.reviews.averageReceivedRating)}
                                />
                                <StatLine label="Написано отзывов" value={formatNumber(dealsUserStats.reviews.written)} />
                              </Stack>
                            </CardContent>
                          </Card>
                        </Grid>

                        <Grid size={{ xs: 12, xl: 4 }}>
                          <Card variant="outlined" sx={{ height: "100%", borderRadius: 2 }}>
                            <CardContent>
                              <Stack spacing={1.5}>
                                <Typography variant="h6" fontWeight={800}>
                                  Жалобы
                                </Typography>
                                <StatLine label="Подал жалоб" value={formatNumber(dealsUserStats.reports.filed)} />
                                <StatLine
                                  label="Принятые жалобы на пользователя"
                                  value={formatNumber(dealsUserStats.reports.received.accepted)}
                                />
                                <StatLine
                                  label="Отклонённые жалобы на пользователя"
                                  value={formatNumber(dealsUserStats.reports.received.rejected)}
                                />
                              </Stack>
                            </CardContent>
                          </Card>
                        </Grid>
                      </Grid>
                    </>
                  ) : null}
                </Stack>
              )}
            </Grid>
          </Grid>
        </Stack>
      </Box>

      <Card variant="outlined" sx={{ borderRadius: 2 }}>
        <CardContent>
          <Stack spacing={2}>
            <Box display="flex" alignItems="center" gap={1}>
              <TagOutlinedIcon color="primary" />
              <Typography variant="h6" fontWeight={800}>
                Теги объявлений
              </Typography>
            </Box>
            {isTagsLoading ? (
              <CircularProgress size={24} />
            ) : tags.length === 0 ? (
              <Alert severity="info">Тегов пока нет.</Alert>
            ) : (
              <Box display="flex" gap={1} flexWrap="wrap">
                {tags.map((tag) => (
                  <Chip
                    key={tag}
                    label={tag}
                    onDelete={() => void deleteAdminTag(tag)}
                    disabled={isDeletingTag}
                    variant="outlined"
                  />
                ))}
              </Box>
            )}
          </Stack>
        </CardContent>
      </Card>

      {isPlatformLoading ? (
        <Box display="flex" justifyContent="center" py={1}>
          <Chip icon={<VisibilityOutlinedIcon />} label="Платформенная статистика обновляется" />
        </Box>
      ) : null}
    </Stack>
  );
}

export default AdminPage;
