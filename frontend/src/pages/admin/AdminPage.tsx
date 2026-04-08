import { Link as RouterLink } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Divider,
  Grid,
  Stack,
  Typography,
} from "@mui/material";
import AdminPanelSettingsOutlinedIcon from "@mui/icons-material/AdminPanelSettingsOutlined";
import Inventory2OutlinedIcon from "@mui/icons-material/Inventory2Outlined";
import SettingsEthernetOutlinedIcon from "@mui/icons-material/SettingsEthernetOutlined";
import ShieldOutlinedIcon from "@mui/icons-material/ShieldOutlined";
import TrendingUpOutlinedIcon from "@mui/icons-material/TrendingUpOutlined";
import chatsApi from "@/features/chats/api/chatsApi";
import dealsApi from "@/features/deals/api/dealsApi";
import offersApi from "@/features/offers/api/offersApi";
import usersApi from "@/features/users/api/usersApi";

const quickActions = [
  { label: "Открыть объявления", to: "/offers" },
  { label: "Открыть сделки", to: "/deals" },
  { label: "Открыть черновики", to: "/deals/drafts" },
  { label: "Открыть чаты", to: "/chats" },
];

const futureModules = [
  {
    title: "Управление пользователями",
    description: "Список пользователей, роли, блокировки и ручная проверка профилей.",
    icon: <ShieldOutlinedIcon color="primary" />,
  },
  {
    title: "Модерация контента",
    description: "Проверка объявлений и жалоб с очередью на ручной разбор.",
    icon: <Inventory2OutlinedIcon color="primary" />,
  },
  {
    title: "Сделки и споры",
    description: "Просмотр проблемных сделок, ручные статусы и журнал решений.",
    icon: <SettingsEthernetOutlinedIcon color="primary" />,
  },
  {
    title: "Метрики",
    description: "Ключевые показатели, воронка активности и динамика платформы.",
    icon: <TrendingUpOutlinedIcon color="primary" />,
  },
];

interface MetricCardProps {
  title: string;
  value: string;
  caption: string;
  isLoading?: boolean;
}

function MetricCard({ title, value, caption, isLoading = false }: MetricCardProps) {
  return (
    <Card
      variant="outlined"
      sx={{
        height: "100%",
        borderRadius: 3,
        background:
          "linear-gradient(180deg, rgba(25,118,210,0.06) 0%, rgba(25,118,210,0.01) 100%)",
      }}
    >
      <CardContent>
        <Typography variant="body2" color="text.secondary" mb={1}>
          {title}
        </Typography>
        {isLoading ? (
          <CircularProgress size={24} />
        ) : (
          <Typography variant="h4" fontWeight={700} mb={1}>
            {value}
          </Typography>
        )}
        <Typography variant="body2" color="text.secondary">
          {caption}
        </Typography>
      </CardContent>
    </Card>
  );
}

function AdminPage() {
  const { data: currentUser, isLoading: isUserLoading } = usersApi.useGetCurrentUserQuery();
  const { data: visibleUsers, isLoading: isVisibleUsersLoading } = chatsApi.useListUsersQuery();
  const { data: chats, isLoading: isChatsLoading } = chatsApi.useListChatsQuery();
  const { data: offers, isLoading: isOffersLoading } = offersApi.useGetOffersQuery({
    sort: "ByTime",
    cursor_limit: 20,
  });
  const { data: deals, isLoading: isDealsLoading } = dealsApi.useGetDealsQuery();
  const { data: drafts, isLoading: isDraftsLoading } = dealsApi.useGetMyDraftDealsQuery();

  const isOverviewLoading =
    isVisibleUsersLoading || isChatsLoading || isOffersLoading || isDealsLoading || isDraftsLoading;

  return (
    <Stack spacing={3}>
      <Box
        sx={{
          p: { xs: 3, md: 4 },
          borderRadius: 4,
          color: "common.white",
          background:
            "radial-gradient(circle at top left, rgba(255,255,255,0.2), transparent 35%), linear-gradient(135deg, #16324f 0%, #1e5b88 55%, #3c89b8 100%)",
        }}
      >
        <Stack spacing={2}>
          <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2} flexWrap="wrap">
            <Box>
              <Box display="flex" alignItems="center" gap={1} mb={1}>
                <AdminPanelSettingsOutlinedIcon />
                <Typography variant="overline" sx={{ opacity: 0.9 }}>
                  Admin Workspace
                </Typography>
              </Box>
              <Typography variant="h3" fontWeight={800} mb={1}>
                Базовая админка
              </Typography>
              <Typography variant="body1" sx={{ maxWidth: 720, opacity: 0.92 }}>
                Это стартовый frontend-каркас для администрирования. Отдельных backend-функций и
                проверки админ-ролей пока нет, поэтому экран показывает только доступные сейчас
                данные и точки расширения.
              </Typography>
            </Box>
            <Chip
              label="Frontend only"
              sx={{
                bgcolor: "rgba(255,255,255,0.14)",
                color: "common.white",
                borderRadius: 2,
                fontWeight: 600,
              }}
            />
          </Box>

          <Box display="flex" gap={1.5} flexWrap="wrap">
            {quickActions.map((action) => (
              <Button
                key={action.to}
                component={RouterLink}
                to={action.to}
                variant="contained"
                sx={{
                  bgcolor: "rgba(255,255,255,0.12)",
                  backdropFilter: "blur(10px)",
                  "&:hover": { bgcolor: "rgba(255,255,255,0.2)" },
                }}
              >
                {action.label}
              </Button>
            ))}
          </Box>
        </Stack>
      </Box>

      <Alert severity="info">
        Страница подготовлена как база для дальнейшего развития: можно без изменения структуры
        добавить реальные админ-эндпоинты, RBAC и отдельные сценарии модерации.
      </Alert>

      <Grid container spacing={2}>
        <Grid size={{ xs: 12, md: 6 }}>
          <Card variant="outlined" sx={{ height: "100%", borderRadius: 3 }}>
            <CardContent>
              <Typography variant="h5" fontWeight={700} mb={2}>
                Текущий доступ
              </Typography>

              {isUserLoading ? (
                <CircularProgress size={24} />
              ) : currentUser ? (
                <Stack spacing={1.5}>
                  <Box>
                    <Typography variant="caption" color="text.secondary">
                      Пользователь
                    </Typography>
                    <Typography variant="body1" fontWeight={600}>
                      {currentUser.name || "Без имени"}
                    </Typography>
                  </Box>
                  <Box>
                    <Typography variant="caption" color="text.secondary">
                      Email
                    </Typography>
                    <Typography variant="body2">{currentUser.email}</Typography>
                  </Box>
                  <Box display="flex" gap={1} flexWrap="wrap">
                    <Chip label="Текущая сессия" color="primary" variant="outlined" />
                    <Chip label="Отдельная admin-role не реализована" variant="outlined" />
                  </Box>
                </Stack>
              ) : (
                <Alert severity="warning">Не удалось получить данные текущего пользователя.</Alert>
              )}
            </CardContent>
          </Card>
        </Grid>

        <Grid size={{ xs: 12, md: 6 }}>
          <Card variant="outlined" sx={{ height: "100%", borderRadius: 3 }}>
            <CardContent>
              <Typography variant="h5" fontWeight={700} mb={2}>
                Что уже можно расширять
              </Typography>
              <Stack spacing={1.5}>
                <Typography variant="body2" color="text.secondary">
                  Маршрут `/admin` уже встроен в основную навигацию и может стать входной точкой
                  для полноценных админ-разделов.
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Карточки ниже используют существующие frontend query-хуки, поэтому сюда можно
                  постепенно подмешивать реальные системные метрики.
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Следующий логичный шаг: вынести отдельный layout, секции и проверку прав доступа.
                </Typography>
              </Stack>
            </CardContent>
          </Card>
        </Grid>
      </Grid>

      <Box>
        <Box display="flex" justifyContent="space-between" alignItems="center" mb={2} gap={2} flexWrap="wrap">
          <Typography variant="h5" fontWeight={700}>
            Обзор текущих данных
          </Typography>
          {isOverviewLoading && <Chip label="Обновление" size="small" />}
        </Box>
        <Grid container spacing={2}>
          <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
            <MetricCard
              title="Пользователи"
              value={String(visibleUsers?.length ?? 0)}
              caption="Количество пользователей, доступных через текущий API."
              isLoading={isVisibleUsersLoading}
            />
          </Grid>
          <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
            <MetricCard
              title="Чаты"
              value={String(chats?.length ?? 0)}
              caption="Диалоги, которые уже можно использовать в будущем центре поддержки."
              isLoading={isChatsLoading}
            />
          </Grid>
          <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
            <MetricCard
              title="Объявления"
              value={String(offers?.offers.length ?? 0)}
              caption="Первые 20 объявлений из текущего списка, без отдельной админ-фильтрации."
              isLoading={isOffersLoading}
            />
          </Grid>
          <Grid size={{ xs: 12, sm: 6, lg: 3 }}>
            <MetricCard
              title="Сделки и черновики"
              value={`${deals?.length ?? 0} / ${drafts?.length ?? 0}`}
              caption="Слева сделки, справа черновики; формат удобен как базовый health snapshot."
              isLoading={isDealsLoading || isDraftsLoading}
            />
          </Grid>
        </Grid>
      </Box>

      <Card variant="outlined" sx={{ borderRadius: 3 }}>
        <CardContent>
          <Typography variant="h5" fontWeight={700} mb={2}>
            Каркас модулей
          </Typography>
          <Grid container spacing={2}>
            {futureModules.map((module) => (
              <Grid key={module.title} size={{ xs: 12, md: 6 }}>
                <Card
                  variant="outlined"
                  sx={{
                    height: "100%",
                    borderRadius: 3,
                    borderStyle: "dashed",
                  }}
                >
                  <CardContent>
                    <Box display="flex" alignItems="center" gap={1.5} mb={1.5}>
                      {module.icon}
                      <Typography variant="h6" fontWeight={700}>
                        {module.title}
                      </Typography>
                    </Box>
                    <Typography variant="body2" color="text.secondary">
                      {module.description}
                    </Typography>
                  </CardContent>
                </Card>
              </Grid>
            ))}
          </Grid>
        </CardContent>
      </Card>

      <Card variant="outlined" sx={{ borderRadius: 3 }}>
        <CardContent>
          <Typography variant="h5" fontWeight={700} mb={2}>
            Технические заметки
          </Typography>
          <Divider sx={{ mb: 2 }} />
          <Stack spacing={1.5}>
            <Typography variant="body2" color="text.secondary">
              1. Сейчас это обычный frontend-маршрут внутри пользовательского приложения.
            </Typography>
            <Typography variant="body2" color="text.secondary">
              2. Доступ не ограничен отдельной ролью, потому что бэкенд ещё не отдаёт такие права.
            </Typography>
            <Typography variant="body2" color="text.secondary">
              3. Структура страницы уже готова для подключения реальных списков, фильтров, таблиц
              и действий администратора.
            </Typography>
          </Stack>
        </CardContent>
      </Card>
    </Stack>
  );
}

export default AdminPage;
