import { useEffect, useMemo } from "react";
import { Link as RouterLink, useSearchParams } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Stack,
  Typography,
} from "@mui/material";
import dealsApi from "@/features/deals/api/dealsApi.ts";
import reviewsApi from "@/features/reviews/api/reviewsApi.ts";
import PendingReviewCard from "@/widgets/reviews/PendingReviewCard.tsx";
import { useAppDispatch, useAppSelector } from "@/hooks/redux.ts";

function PendingReviewsPage() {
  const dispatch = useAppDispatch();
  const [searchParams] = useSearchParams();
  const selectedDealId = searchParams.get("dealId");

  const { data: deals, isLoading, error } = dealsApi.useGetDealsQuery({ my: true });

  const completedDeals = useMemo(
    () => (deals ?? []).filter((deal) => deal.status === "Completed"),
    [deals],
  );

  useEffect(() => {
    if (completedDeals.length === 0) {
      return;
    }

    const subscriptions = completedDeals.map((deal) =>
      dispatch(reviewsApi.endpoints.getDealPendingReviews.initiate(deal.id)),
    );

    return () => {
      subscriptions.forEach((subscription) => subscription.unsubscribe());
    };
  }, [completedDeals, dispatch]);

  const pendingByDeal = useAppSelector((state) =>
    completedDeals.map((deal) => ({
      deal,
      query: reviewsApi.endpoints.getDealPendingReviews.select(deal.id)(state),
    })),
  );

  const pendingReviews = useMemo(
    () =>
      pendingByDeal.flatMap(({ deal, query }) =>
        (query.data ?? [])
          .filter((review) => review.canCreate)
          .map((review) => ({ deal, review })),
      ),
    [pendingByDeal],
  );

  const filteredReviews = useMemo(
    () =>
      selectedDealId
        ? pendingReviews.filter(({ deal }) => deal.id === selectedDealId)
        : pendingReviews,
    [pendingReviews, selectedDealId],
  );

  const isPendingLoading =
    completedDeals.length > 0 &&
    pendingByDeal.some(({ query }) => query.isLoading || query.isUninitialized);
  const hasPendingErrors = pendingByDeal.some(({ query }) => query.isError);

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return <Alert severity="error">Не удалось загрузить завершенные сделки</Alert>;
  }

  return (
    <Box maxWidth={900} mx="auto">
      <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2} flexWrap="wrap" mb={3}>
        <div>
          <Typography variant="h4" fontWeight={700}>
            Отзывы после сделки
          </Typography>
          <Typography variant="body1" color="text.secondary" mt={1}>
            Здесь собраны все позиции, по которым вы можете оставить отзыв как получатель.
          </Typography>
        </div>

        <Box display="flex" gap={1} flexWrap="wrap">
          <Button component={RouterLink} to="/reviews/mine" variant="outlined">
            Мои отзывы
          </Button>
          <Button component={RouterLink} to="/profile" variant="text">
            Профиль
          </Button>
        </Box>
      </Box>

      {selectedDealId && (
        <Alert
          severity="info"
          action={
            <Button component={RouterLink} to="/reviews/pending" color="inherit" size="small">
              Все сделки
            </Button>
          }
          sx={{ mb: 3 }}
        >
          Показаны отзывы только по одной сделке.
        </Alert>
      )}

      {hasPendingErrors && (
        <Alert severity="warning" sx={{ mb: 3 }}>
          Часть завершенных сделок не удалось проверить на доступные отзывы.
        </Alert>
      )}

      {isPendingLoading ? (
        <Box display="flex" justifyContent="center" py={6}>
          <CircularProgress />
        </Box>
      ) : completedDeals.length === 0 ? (
        <Alert severity="info">У вас пока нет завершенных сделок.</Alert>
      ) : filteredReviews.length === 0 ? (
        <Alert severity="success">
          Доступных для отзыва позиций сейчас нет. Возможно, вы уже оставили все отзывы.
        </Alert>
      ) : (
        <Stack spacing={2}>
          {filteredReviews.map(({ deal, review }) => (
            <PendingReviewCard
              key={`${deal.id}:${review.itemRef?.itemId ?? review.offerRef?.offerId ?? review.contextType}`}
              dealName={deal.name ?? "Сделка"}
              review={review}
            />
          ))}
        </Stack>
      )}
    </Box>
  );
}

export default PendingReviewsPage;
