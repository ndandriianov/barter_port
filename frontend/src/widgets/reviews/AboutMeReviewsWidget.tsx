import {
  Alert,
  Box,
  CircularProgress,
  Stack,
} from "@mui/material";
import reviewsApi from "@/features/reviews/api/reviewsApi.ts";
import usersApi from "@/features/users/api/usersApi.ts";
import ReviewCard from "@/widgets/reviews/ReviewCard.tsx";
import ReviewSummaryCard from "@/widgets/reviews/ReviewSummaryCard.tsx";

function AboutMeReviewsWidget() {
  const { data: me, isLoading: isMeLoading, error: meError } = usersApi.useGetCurrentUserQuery();
  const {
    data: reviews,
    isLoading,
    error,
  } = reviewsApi.useGetProviderReviewsQuery(me?.id ?? "", {
    skip: !me?.id,
  });
  const { data: summary, isLoading: isSummaryLoading } = reviewsApi.useGetProviderReviewsSummaryQuery(me?.id ?? "", {
    skip: !me?.id,
  });

  if (isMeLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (meError || !me) {
    return <Alert severity="error">Не удалось определить текущего пользователя</Alert>;
  }

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return <Alert severity="error">Не удалось загрузить отзывы о вас</Alert>;
  }

  return (
    <Box>
      {summary && !isSummaryLoading && (
        <Box mb={3}>
          <ReviewSummaryCard title="Сводка по оценкам" summary={summary} />
        </Box>
      )}

      {!reviews || reviews.length === 0 ? (
        <Alert severity="info">Отзывов о вас пока нет.</Alert>
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

export default AboutMeReviewsWidget;
