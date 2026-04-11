import { useMemo, useState } from "react";
import { Link as RouterLink } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  Rating,
  Stack,
  Typography,
} from "@mui/material";
import type { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import type { SerializedError } from "@reduxjs/toolkit";
import dealsApi from "@/features/deals/api/dealsApi.ts";
import offersApi from "@/features/offers/api/offersApi.ts";
import reviewsApi from "@/features/reviews/api/reviewsApi.ts";
import usersApi from "@/features/users/api/usersApi.ts";
import { formatReviewDate, reviewContextLabels } from "@/features/reviews/model/meta.ts";
import type { Review } from "@/features/reviews/model/types.ts";
import type { OfferType } from "@/features/offers/model/types.ts";
import { getStatusCode } from "@/shared/utils/getStatusCode.ts";
import ReviewEditorDialog from "@/widgets/reviews/ReviewEditorDialog.tsx";

interface ReviewCardProps {
  review: Review;
  editable?: boolean;
}

const subjectTypeLabel: Record<OfferType, string> = {
  good: "Товар",
  service: "Услуга",
};

function getReviewMutationError(
  error: FetchBaseQueryError | SerializedError | undefined,
  fallback: string,
): string | null {
  if (!error) {
    return null;
  }

  const code = getStatusCode(error);
  switch (code) {
    case 400:
      return "Проверьте оценку и текст отзыва";
    case 403:
      return "Вы не можете изменить этот отзыв";
    case 404:
      return "Отзыв больше недоступен";
    default:
      return fallback;
  }
}

function ReviewCard({ review, editable = false }: ReviewCardProps) {
  const [isEditOpen, setIsEditOpen] = useState(false);
  const [updateReview, { isLoading: isUpdating, error: updateError }] = reviewsApi.useUpdateReviewMutation();
  const [deleteReview, { isLoading: isDeleting, error: deleteError }] = reviewsApi.useDeleteReviewMutation();
  const { data: author } = usersApi.useGetUserByIdQuery(review.authorId);
  const { data: provider } = usersApi.useGetUserByIdQuery(review.providerId);
  const { data: deal } = dealsApi.useGetDealByIdQuery(review.itemRef?.dealId ?? review.dealId, {
    skip: !review.itemRef,
  });
  const { data: offer } = offersApi.useGetOfferByIdQuery(review.offerRef?.offerId ?? "", {
    skip: !review.offerRef,
  });

  const contextLabel = useMemo(() => {
    if (review.offerRef && review.itemRef) {
      return reviewContextLabels["offer+item"];
    }
    if (review.offerRef) {
      return reviewContextLabels["offer-only"];
    }
    return reviewContextLabels["item-only"];
  }, [review.itemRef, review.offerRef]);

  const subject = useMemo(() => {
    if (review.itemRef && deal) {
      const item = deal.items.find((dealItem) => dealItem.id === review.itemRef?.itemId);
      if (item?.name?.trim()) {
        return {
          label: subjectTypeLabel[item.type],
          name: item.name.trim(),
        };
      }
    }

    if (offer?.name?.trim()) {
      return {
        label: subjectTypeLabel[offer.type],
        name: offer.name.trim(),
      };
    }

    return null;
  }, [deal, offer, review.itemRef]);

  const handleUpdate = async ({ rating, comment }: { rating: number; comment: string }) => {
    const body: { rating?: number; comment?: string } = {};

    if (rating !== review.rating) {
      body.rating = rating;
    }

    if (comment !== (review.comment ?? "")) {
      body.comment = comment;
    }

    if (Object.keys(body).length === 0) {
      setIsEditOpen(false);
      return;
    }

    await updateReview({
      reviewId: review.id,
      body,
    }).unwrap();

    setIsEditOpen(false);
  };

  const handleDelete = async () => {
    if (!window.confirm("Удалить этот отзыв?")) {
      return;
    }

    await deleteReview(review.id).unwrap();
  };

  return (
    <>
      <Card variant="outlined">
        <CardContent>
          <Box display="flex" justifyContent="space-between" gap={2} flexWrap="wrap" mb={1.5}>
            <Stack spacing={0.5}>
              <Typography variant="subtitle1" fontWeight={700}>
                {provider?.name?.trim() || "Поставщик"}
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Автор: {author?.name?.trim() || "Имя не указано"}
              </Typography>
            </Stack>

            <Box textAlign={{ xs: "left", sm: "right" }}>
              <Rating value={review.rating} readOnly />
              <Typography variant="caption" color="text.secondary" display="block">
                {formatReviewDate(review.updatedAt ?? review.createdAt)}
              </Typography>
            </Box>
          </Box>

          <Box display="flex" gap={1} flexWrap="wrap" mb={1.5}>
            <Chip label={contextLabel} size="small" variant="outlined" />
            <Button component={RouterLink} to={`/deals/${review.dealId}`} size="small">
              Сделка
            </Button>
            <Button component={RouterLink} to={`/users/${review.providerId}/reviews`} size="small">
              Отзывы о поставщике
            </Button>
            {review.offerRef && (
              <Button component={RouterLink} to={`/offers/${review.offerRef.offerId}`} size="small">
                Offer
              </Button>
            )}
            {review.itemRef && (
              <Button
                component={RouterLink}
                to={`/deals/${review.itemRef.dealId}/items/${review.itemRef.itemId}`}
                size="small"
              >
                Позиция
              </Button>
            )}
          </Box>

          {subject && (
            <Typography variant="body2" fontWeight={600} mb={1}>
              {subject.label}: {subject.name}
            </Typography>
          )}

          <Typography variant="body2" color={review.comment ? "text.primary" : "text.secondary"}>
            {review.comment?.trim() ? review.comment : "Комментарий не добавлен"}
          </Typography>

          {(editable || deleteError) && (
            <Box display="flex" gap={1} mt={2} flexWrap="wrap">
              {editable && (
                <Button variant="outlined" size="small" onClick={() => setIsEditOpen(true)}>
                  Редактировать
                </Button>
              )}
              {editable && (
                <Button
                  variant="outlined"
                  color="error"
                  size="small"
                  onClick={() => void handleDelete()}
                  disabled={isDeleting}
                >
                  Удалить
                </Button>
              )}
            </Box>
          )}

          {deleteError && (
            <Alert severity="error" sx={{ mt: 1.5 }}>
              {getReviewMutationError(deleteError, "Не удалось удалить отзыв")}
            </Alert>
          )}
        </CardContent>
      </Card>

      <ReviewEditorDialog
        open={isEditOpen}
        title="Редактировать отзыв"
        submitLabel="Сохранить"
        initialRating={review.rating}
        initialComment={review.comment ?? ""}
        isLoading={isUpdating}
        errorMessage={getReviewMutationError(updateError, "Не удалось обновить отзыв")}
        onClose={() => setIsEditOpen(false)}
        onSubmit={handleUpdate}
      />
    </>
  );
}

export default ReviewCard;
