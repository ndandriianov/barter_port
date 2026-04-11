import { Link as RouterLink, useParams } from "react-router-dom";
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
import ReviewSummaryCard from "@/widgets/reviews/ReviewSummaryCard.tsx";

function UserReviewsPage() {
  const { userId } = useParams<{ userId: string }>();
  const { data: currentUser } = usersApi.useGetCurrentUserQuery();
  const { data: user } = usersApi.useGetUserByIdQuery(userId ?? "", {
    skip: !userId,
  });
  const {
    data: reviews,
    isLoading,
    error,
  } = reviewsApi.useGetProviderReviewsQuery(userId ?? "", {
    skip: !userId,
  });
  const { data: summary, isLoading: isSummaryLoading } = reviewsApi.useGetProviderReviewsSummaryQuery(userId ?? "", {
    skip: !userId,
  });

  if (!userId) {
    return <Alert severity="warning">Пользователь не найден</Alert>;
  }

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return <Alert severity="error">Не удалось загрузить отзывы о пользователе</Alert>;
  }

  const isCurrentUser = currentUser?.id === userId;
  const title = isCurrentUser
    ? "Отзывы обо мне"
    : `Отзывы о пользователе${user?.name?.trim() ? ` ${user.name.trim()}` : ""}`;

  return (
    <Box maxWidth={900} mx="auto">
      <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2} flexWrap="wrap" mb={3}>
        <div>
          <Typography variant="h4" fontWeight={700}>
            {title}
          </Typography>
          <Typography variant="body1" color="text.secondary" mt={1}>
            Отзывы показывают опыт других участников сделок с этим поставщиком.
          </Typography>
        </div>

        <Box display="flex" gap={1} flexWrap="wrap">
          {isCurrentUser && (
            <Button component={RouterLink} to="/reviews/mine" variant="outlined">
              Мои отзывы
            </Button>
          )}
          <Button component={RouterLink} to="/reviews/pending" variant="text">
            Доступные отзывы
          </Button>
        </Box>
      </Box>

      {summary && !isSummaryLoading && (
        <Box mb={3}>
          <ReviewSummaryCard title="Сводка по оценкам" summary={summary} />
        </Box>
      )}

      {!reviews || reviews.length === 0 ? (
        <Alert severity="info">Отзывов о пользователе пока нет.</Alert>
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

export default UserReviewsPage;
