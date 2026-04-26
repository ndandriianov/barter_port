import { useMemo } from "react";
import { Alert, Box, Button, Card, CardContent, Chip, Grid, Stack, Typography } from "@mui/material";
import { Link as RouterLink } from "react-router-dom";
import usersApi from "@/features/users/api/usersApi.ts";
import { appRoutes } from "@/shared/config/appRoutes.ts";
import ProfileSectionShell from "@/widgets/profile/ProfileSectionShell.tsx";

interface ProfileReputationPageProps {
  historyMode?: boolean;
}

function formatSourceType(sourceType: string) {
  switch (sourceType) {
    case "deals.offer_report.penalty":
      return "Штраф по жалобе на объявление";
    case "deals.deal_failure.responsible":
      return "Штраф за провал сделки";
    case "deals.deal_completion.reward":
      return "Завершение сделки";
    case "deals.review_creation.reward":
      return "Оставленный отзыв";
    default:
      return sourceType;
  }
}

function ProfileReputationPage({ historyMode = false }: ProfileReputationPageProps) {
  const { data: me } = usersApi.useGetCurrentUserQuery();
  const {
    data: reputationEvents,
    isLoading,
    error,
    refetch,
    isFetching,
  } = usersApi.useGetCurrentUserReputationEventsQuery();

  const positiveCount = useMemo(
    () => (reputationEvents ?? []).filter((event) => event.delta > 0).length,
    [reputationEvents],
  );
  const negativeCount = useMemo(
    () => (reputationEvents ?? []).filter((event) => event.delta < 0).length,
    [reputationEvents],
  );
  const totalDelta = useMemo(
    () => (reputationEvents ?? []).reduce((sum, event) => sum + event.delta, 0),
    [reputationEvents],
  );
  const visibleEvents = historyMode ? reputationEvents ?? [] : (reputationEvents ?? []).slice(0, 5);

  return (
    <ProfileSectionShell
      title={historyMode ? "История репутации" : "Репутация"}
      description=""
      actions={
        <Button variant="outlined" onClick={() => refetch()} disabled={isFetching}>
          Обновить
        </Button>
      }
    >
      <Stack spacing={3}>
        <Grid container spacing={2.5}>
          <Grid size={{ xs: 12, md: 4 }}>
            <Card variant="outlined">
              <CardContent>
                <Typography variant="caption" color="text.secondary">
                  Текущая репутация
                </Typography>
                <Typography variant="h3" fontWeight={900} color="primary.main">
                  {me?.reputationPoints ?? 0}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid size={{ xs: 12, md: 4 }}>
            <Card variant="outlined">
              <CardContent>
                <Typography variant="caption" color="text.secondary">
                  Позитивные события
                </Typography>
                <Typography variant="h4" fontWeight={800} color="success.main">
                  {positiveCount}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
          <Grid size={{ xs: 12, md: 4 }}>
            <Card variant="outlined">
              <CardContent>
                <Typography variant="caption" color="text.secondary">
                  Негативные события
                </Typography>
                <Typography variant="h4" fontWeight={800} color="error.main">
                  {negativeCount}
                </Typography>
              </CardContent>
            </Card>
          </Grid>
        </Grid>

        <Alert severity="info">
          Суммарная дельта по событиям: <strong>{totalDelta >= 0 ? `+${totalDelta}` : totalDelta}</strong>. Репутация
          формируется завершением сделок, публикацией отзывов и штрафами по модерации/провалам.
        </Alert>

        <Box display="flex" gap={1.5} flexWrap="wrap">
          <Button
            component={RouterLink}
            to={appRoutes.profile.reputation}
            variant={!historyMode ? "contained" : "outlined"}
          >
            Обзор
          </Button>
          <Button
            component={RouterLink}
            to={appRoutes.profile.reputationHistory}
            variant={historyMode ? "contained" : "outlined"}
          >
            История событий
          </Button>
        </Box>

        {isLoading ? (
          <Typography color="text.secondary">Загрузка репутационных событий...</Typography>
        ) : error ? (
          <Alert severity="error">Не удалось загрузить историю репутации.</Alert>
        ) : !reputationEvents || reputationEvents.length === 0 ? (
          <Alert severity="info">История изменения репутации пока пуста.</Alert>
        ) : (
          <Stack spacing={1.5}>
            {visibleEvents.map((event) => (
              <Card key={event.id} variant="outlined">
                <CardContent>
                  <Stack spacing={0.75}>
                    <Box display="flex" justifyContent="space-between" alignItems="center" gap={2} flexWrap="wrap">
                      <Typography variant="body2" color="text.secondary">
                        {new Date(event.createdAt).toLocaleString("ru-RU")}
                      </Typography>
                      <Chip
                        label={event.delta >= 0 ? `+${event.delta}` : `${event.delta}`}
                        color={event.delta >= 0 ? "success" : "error"}
                        size="small"
                      />
                    </Box>
                    <Typography variant="subtitle1" fontWeight={700}>
                      {formatSourceType(event.sourceType)}
                    </Typography>
                    {event.comment && (
                      <Typography variant="body2" color="text.secondary">
                        {event.comment}
                      </Typography>
                    )}
                  </Stack>
                </CardContent>
              </Card>
            ))}
          </Stack>
        )}

        {!historyMode && reputationEvents && reputationEvents.length > 5 && (
          <Box>
            <Button component={RouterLink} to={appRoutes.profile.reputationHistory} variant="outlined">
              Открыть полную историю
            </Button>
          </Box>
        )}
      </Stack>
    </ProfileSectionShell>
  );
}

export default ProfileReputationPage;
