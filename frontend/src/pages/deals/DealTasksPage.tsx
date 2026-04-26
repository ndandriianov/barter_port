import { Alert, Box, Button, Card, CardContent, Chip, Stack, Typography } from "@mui/material";
import { Link as RouterLink } from "react-router-dom";
import PendingReviewCard from "@/widgets/reviews/PendingReviewCard.tsx";
import dealsApi from "@/features/deals/api/dealsApi.ts";
import useDealActionQueue from "@/features/deals/model/useDealActionQueue.ts";
import { appRoutes } from "@/shared/config/appRoutes.ts";

function DealTasksPage() {
  const { incomingDrafts, dealsWithJoinRequests, pendingReviews, totalActionCount, isLoading, error } =
    useDealActionQueue();
  const { data: deals = [] } = dealsApi.useGetDealsQuery({ my: true });

  const dealNameById = new Map(deals.map((deal) => [deal.id, deal.name ?? `Сделка ${deal.id}`]));

  if (isLoading) {
    return <Typography color="text.secondary">Загрузка очереди действий...</Typography>;
  }

  if (error) {
    return <Alert severity="error">Не удалось собрать очередь действий по сделкам.</Alert>;
  }

  return (
    <Stack spacing={3}>
      <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2} flexWrap="wrap">
        <Box>
          <Typography variant="overline" color="warning.main">
            Сделки / Нужны действия
          </Typography>
          <Typography variant="h4" fontWeight={800} mb={1}>
            Очередь задач
          </Typography>
        </Box>
        <Chip label={`${totalActionCount} задач`} color="warning" />
      </Box>

      {totalActionCount === 0 && (
        <Alert severity="success">
          Срочных действий по сделкам сейчас нет. Можно перейти к активным сделкам или истории.
        </Alert>
      )}

      <Card variant="outlined">
        <CardContent>
          <Stack spacing={2}>
            <Box display="flex" justifyContent="space-between" gap={2} flexWrap="wrap">
              <div>
                <Typography variant="h6" fontWeight={700}>
                  Входящие черновики
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Подтвердить участие либо отклонить
                </Typography>
              </div>
              <Button component={RouterLink} to={appRoutes.deals.draftsIncoming} variant="outlined">
                Все входящие
              </Button>
            </Box>

            {incomingDrafts.length === 0 ? (
              <Typography color="text.secondary">Новых входящих черновиков нет.</Typography>
            ) : (
              <Stack spacing={1.25}>
                {incomingDrafts.map((draft) => (
                  <Card key={draft.id} variant="outlined" sx={{ bgcolor: "background.default" }}>
                    <CardContent>
                      <Box display="flex" justifyContent="space-between" alignItems="center" gap={2} flexWrap="wrap">
                        <Box>
                          <Typography fontWeight={700}>{draft.name ?? `Черновик ${draft.id}`}</Typography>
                          <Typography variant="body2" color="text.secondary">
                            Участников: {draft.participants.length}
                          </Typography>
                        </Box>
                        <Button component={RouterLink} to={appRoutes.deals.draftDetail(draft.id)} variant="contained">
                          Открыть
                        </Button>
                      </Box>
                    </CardContent>
                  </Card>
                ))}
              </Stack>
            )}
          </Stack>
        </CardContent>
      </Card>

      <Card variant="outlined">
        <CardContent>
          <Stack spacing={2}>
            <Box display="flex" justifyContent="space-between" gap={2} flexWrap="wrap">
              <div>
                <Typography variant="h6" fontWeight={700}>
                  Заявки на присоединение
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Сделки, где есть активные заявки на присоединение и нужно принять решение
                </Typography>
              </div>
              <Button component={RouterLink} to={appRoutes.deals.active} variant="outlined">
                Активные сделки
              </Button>
            </Box>

            {dealsWithJoinRequests.length === 0 ? (
              <Typography color="text.secondary">Необработанных заявок на присоединение сейчас нет.</Typography>
            ) : (
              <Stack spacing={1.25}>
                {dealsWithJoinRequests.map(({ deal, requests }) => (
                  <Card key={deal.id} variant="outlined" sx={{ bgcolor: "background.default" }}>
                    <CardContent>
                      <Box display="flex" justifyContent="space-between" alignItems="center" gap={2} flexWrap="wrap">
                        <Box>
                          <Typography fontWeight={700}>{deal.name ?? `Сделка ${deal.id}`}</Typography>
                          <Typography variant="body2" color="text.secondary">
                            Ожидают решения: {requests.length}
                          </Typography>
                        </Box>
                        <Button component={RouterLink} to={appRoutes.deals.detail(deal.id)} variant="contained">
                          Открыть сделку
                        </Button>
                      </Box>
                    </CardContent>
                  </Card>
                ))}
              </Stack>
            )}
          </Stack>
        </CardContent>
      </Card>

      <Stack spacing={2}>
        <Box display="flex" justifyContent="space-between" gap={2} flexWrap="wrap">
          <div>
            <Typography variant="h6" fontWeight={700}>
              Оставить отзыв
            </Typography>
            <Typography variant="body2" color="text.secondary">
              После завершения сделки можно оставить отзыв на полученные товары
            </Typography>
          </div>
        </Box>

        {pendingReviews.length === 0 ? (
          <Alert severity="info">Доступных отзывов для публикации сейчас нет.</Alert>
        ) : (
          pendingReviews.map(({ deal, review }) => (
            <PendingReviewCard
              key={`${deal.id}-${review.itemRef?.itemId ?? review.offerRef?.offerId ?? review.providerId ?? "review"}`}
              review={review}
              dealName={dealNameById.get(deal.id)}
            />
          ))
        )}
      </Stack>
    </Stack>
  );
}

export default DealTasksPage;
