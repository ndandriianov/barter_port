import {
  Alert,
  Box,
  CircularProgress,
  Stack,
} from "@mui/material";
import reviewsApi from "@/features/reviews/api/reviewsApi.ts";
import usersApi from "@/features/users/api/usersApi.ts";
import ReviewCard from "@/widgets/reviews/ReviewCard.tsx";

function MyReviewsWidget() {
  const { data: me, isLoading: isMeLoading, error: meError } = usersApi.useGetCurrentUserQuery();
  const {
    data: reviews,
    isLoading,
    error,
  } = reviewsApi.useGetAuthorReviewsQuery(me?.id ?? "", {
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
    return <Alert severity="error">Не удалось загрузить ваши отзывы</Alert>;
  }

  if (!reviews || reviews.length === 0) {
    return <Alert severity="info">Вы еще не оставляли отзывов.</Alert>;
  }

  return (
    <Stack spacing={2}>
      {reviews.map((review) => (
        <ReviewCard key={review.id} review={review} editable />
      ))}
    </Stack>
  );
}

export default MyReviewsWidget;
