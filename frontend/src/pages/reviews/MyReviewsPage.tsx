import { Link as RouterLink } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  CircularProgress,
  Stack,
  Typography,
} from "@mui/material";
import reviewsApi from "@/features/reviews/api/reviewsApi.ts";
import usersApi from "@/features/users/api/usersApi.ts";
import ReviewCard from "@/widgets/reviews/ReviewCard.tsx";

function MyReviewsPage() {
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

  return (
    <Box maxWidth={900} mx="auto">
      <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2} flexWrap="wrap" mb={3}>
        <div>
          <Typography variant="h4" fontWeight={700}>
            Мои отзывы
          </Typography>
          <Typography variant="body1" color="text.secondary" mt={1}>
            Здесь можно посмотреть, отредактировать или удалить отзывы, которые вы оставили после сделок.
          </Typography>
        </div>

        <Box display="flex" gap={1} flexWrap="wrap">
          <Button component={RouterLink} to="/reviews/pending" variant="outlined">
            Доступные отзывы
          </Button>
          <Button component={RouterLink} to={`/users/${me.id}/reviews`} variant="text">
            Отзывы обо мне
          </Button>
        </Box>
      </Box>

      {!reviews || reviews.length === 0 ? (
        <Alert severity="info">Вы еще не оставляли отзывов.</Alert>
      ) : (
        <Stack spacing={2}>
          {reviews.map((review) => (
            <ReviewCard key={review.id} review={review} editable />
          ))}
        </Stack>
      )}
    </Box>
  );
}

export default MyReviewsPage;
