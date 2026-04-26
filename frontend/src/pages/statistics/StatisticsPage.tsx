import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  CircularProgress,
  Divider,
  Grid,
  Stack,
  Typography,
} from "@mui/material";
import { Link as RouterLink } from "react-router-dom";
import statisticsApi from "@/features/statistics/api/statisticsApi";
import { appRoutes } from "@/shared/config/appRoutes.ts";
import ProfileSectionShell from "@/widgets/profile/ProfileSectionShell.tsx";

function StatCard({
  label,
  value,
  color,
}: {
  label: string;
  value: string | number;
  color?: string;
}) {
  return (
    <Box>
      <Typography variant="caption" color="text.secondary" display="block">
        {label}
      </Typography>
      <Typography variant="h4" fontWeight={700} color={color}>
        {value}
      </Typography>
    </Box>
  );
}

function SectionCard({ title, children }: { title: string; children: React.ReactNode }) {
  return (
    <Card variant="outlined">
      <CardContent>
        <Typography variant="h6" fontWeight={700} mb={2}>
          {title}
        </Typography>
        <Divider sx={{ mb: 2 }} />
        {children}
      </CardContent>
    </Card>
  );
}

function StatisticsPage() {
  const { data, isLoading, error, refetch, isFetching } =
    statisticsApi.useGetMyStatisticsQuery();

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={8}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Alert
        severity="error"
        action={
          <Button color="inherit" size="small" onClick={() => refetch()}>
            Повторить
          </Button>
        }
      >
        Не удалось загрузить статистику.
      </Alert>
    );
  }

  if (!data) return null;

  const avgRating =
    data.reviews.averageRatingReceived !== null
      ? data.reviews.averageRatingReceived.toFixed(2)
      : "—";

  return (
    <ProfileSectionShell
      title="Статистика"
      description="Персональные метрики пользователя"
      actions={
        <Button variant="outlined" onClick={() => refetch()} disabled={isFetching}>
          Обновить
        </Button>
      }
    >
      <Box maxWidth={900}>
        <Grid container spacing={3}>
          <Grid size={{ xs: 12, sm: 6 }}>
            <SectionCard title="Сделки">
              <Grid container spacing={2}>
                <Grid size={{ xs: 4 }}>
                  <StatCard
                    label="Завершено"
                    value={data.deals.completed}
                    color="success.main"
                  />
                </Grid>
                <Grid size={{ xs: 4 }}>
                  <StatCard label="Провалено" value={data.deals.failed} color="error.main" />
                </Grid>
                <Grid size={{ xs: 4 }}>
                  <StatCard label="Активных" value={data.deals.active} color="info.main" />
                </Grid>
              </Grid>
            </SectionCard>
          </Grid>

          <Grid size={{ xs: 12, sm: 6 }}>
            <SectionCard title="Объявления">
              <Grid container spacing={2}>
                <Grid size={{ xs: 6 }}>
                  <StatCard label="Всего" value={data.offers.total} />
                </Grid>
                <Grid size={{ xs: 6 }}>
                  <StatCard label="Просмотров" value={data.offers.totalViews} />
                </Grid>
              </Grid>
            </SectionCard>
          </Grid>

          <Grid size={{ xs: 12, sm: 6 }}>
            <SectionCard title="Отзывы">
              <Grid container spacing={2}>
                <Grid size={{ xs: 4 }}>
                  <StatCard label="Написано" value={data.reviews.written} />
                </Grid>
                <Grid size={{ xs: 4 }}>
                  <StatCard label="Получено" value={data.reviews.received} />
                </Grid>
                <Grid size={{ xs: 4 }}>
                  <StatCard label="Средний рейтинг" value={avgRating} color="warning.main" />
                </Grid>
              </Grid>
            </SectionCard>
          </Grid>

          <Grid size={{ xs: 12, sm: 6 }}>
            <SectionCard title="Жалобы">
              <Stack spacing={2.5}>
                <Grid container spacing={2}>
                  <Grid size={{ xs: 6 }}>
                    <StatCard label="Подано мной" value={data.reports.filedByMe} />
                  </Grid>
                  <Grid size={{ xs: 6 }}>
                    <StatCard
                      label="На мои объявления"
                      value={data.reports.onMyOffers.total}
                      color={data.reports.onMyOffers.total > 0 ? "warning.main" : undefined}
                    />
                  </Grid>
                  <Grid size={{ xs: 4 }}>
                    <StatCard
                      label="На модерации"
                      value={data.reports.onMyOffers.pending}
                      color={data.reports.onMyOffers.pending > 0 ? "warning.main" : undefined}
                    />
                  </Grid>
                  <Grid size={{ xs: 4 }}>
                    <StatCard
                      label="Принято"
                      value={data.reports.onMyOffers.accepted}
                      color={data.reports.onMyOffers.accepted > 0 ? "error.main" : undefined}
                    />
                  </Grid>
                  <Grid size={{ xs: 4 }}>
                    <StatCard label="Отклонено" value={data.reports.onMyOffers.rejected} />
                  </Grid>
                </Grid>

                <Alert severity="info">
                  Подробно можно открыть только жалобы на ваши объявления. Для жалоб, которые подали вы сами,
                  backend сейчас отдаёт только счётчик `filedByMe`, без отдельного списка.
                </Alert>

                <Box display="flex" gap={1.5} flexWrap="wrap">
                  <Button
                    component={RouterLink}
                    to={appRoutes.market.myPublicationModeration}
                    variant="outlined"
                    color="warning"
                  >
                    Открыть жалобы на мои объявления
                  </Button>
                </Box>
              </Stack>
            </SectionCard>
          </Grid>
        </Grid>
      </Box>
    </ProfileSectionShell>
  );
}

export default StatisticsPage;
