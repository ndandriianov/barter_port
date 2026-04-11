import { Link as RouterLink } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Stack,
} from "@mui/material";
import PendingReviewCard from "@/widgets/reviews/PendingReviewCard.tsx";
import usePendingReviews from "@/features/reviews/model/usePendingReviews.ts";

interface AvailableReviewsWidgetProps {
  selectedDealId?: string | null;
}

function AvailableReviewsWidget({ selectedDealId }: AvailableReviewsWidgetProps) {
  const {
    completedDeals,
    filteredReviews,
    isDealsLoading,
    dealsError,
    isPendingLoading,
    hasPendingErrors,
  } = usePendingReviews({ selectedDealId });

  if (isDealsLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (dealsError) {
    return <Alert severity="error">Не удалось загрузить завершенные сделки</Alert>;
  }

  return (
    <Box>
      {selectedDealId && (
        <Alert
          severity="info"
          action={
            <Button component={RouterLink} to="/reviews?tab=available" color="inherit" size="small">
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

export default AvailableReviewsWidget;
