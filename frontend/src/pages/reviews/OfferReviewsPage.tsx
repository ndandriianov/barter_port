import { Link as RouterLink, useParams } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Stack,
  Typography,
} from "@mui/material";
import offersApi from "@/features/offers/api/offersApi.ts";
import reviewsApi from "@/features/reviews/api/reviewsApi.ts";
import ReviewCard from "@/widgets/reviews/ReviewCard.tsx";
import ReviewSummaryCard from "@/widgets/reviews/ReviewSummaryCard.tsx";

function OfferReviewsPage() {
  const { offerId } = useParams<{ offerId: string }>();
  const { data: offer, isLoading: isOfferLoading, error: offerError } = offersApi.useGetOfferByIdQuery(offerId ?? "", {
    skip: !offerId,
  });
  const {
    data: reviews,
    isLoading,
    error,
  } = reviewsApi.useGetOfferReviewsQuery(offerId ?? "", {
    skip: !offerId,
  });
  const { data: summary, isLoading: isSummaryLoading } = reviewsApi.useGetOfferReviewsSummaryQuery(offerId ?? "", {
    skip: !offerId,
  });

  if (!offerId) {
    return <Alert severity="warning">Offer не найден</Alert>;
  }

  if (isOfferLoading || isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (offerError || !offer) {
    return <Alert severity="error">Не удалось загрузить offer</Alert>;
  }

  if (error) {
    return <Alert severity="error">Не удалось загрузить отзывы по offer</Alert>;
  }

  return (
    <Box maxWidth={900} mx="auto">
      <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2} flexWrap="wrap" mb={3}>
        <div>
          <Typography variant="h4" fontWeight={700}>
            Отзывы по offer
          </Typography>
          <Typography variant="body1" color="text.secondary" mt={1}>
            {offer.name}
          </Typography>
        </div>

        <Button component={RouterLink} to={`/offers/${offer.id}`} variant="outlined">
          Вернуться к offer
        </Button>
      </Box>

      {summary && !isSummaryLoading && (
        <Box mb={3}>
          <ReviewSummaryCard title="Сводка по этому offer" summary={summary} />
        </Box>
      )}

      {!reviews || reviews.length === 0 ? (
        <Alert severity="info">По этому offer пока нет отзывов.</Alert>
      ) : (
        <Stack spacing={2}>
          {reviews.map((review) => (
            <ReviewCard key={review.id} review={review} />
          ))}
        </Stack>
      )}
    </Box>
  );
}

export default OfferReviewsPage;
