import { useMemo, useState } from "react";
import { Link as RouterLink } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  Stack,
  Typography,
} from "@mui/material";
import type { FetchBaseQueryError } from "@reduxjs/toolkit/query";
import type { SerializedError } from "@reduxjs/toolkit";
import dealsApi from "@/features/deals/api/dealsApi.ts";
import offersApi from "@/features/offers/api/offersApi.ts";
import reviewsApi from "@/features/reviews/api/reviewsApi.ts";
import { reviewContextLabels } from "@/features/reviews/model/meta.ts";
import type { PendingDealReview } from "@/features/reviews/model/types.ts";
import usersApi from "@/features/users/api/usersApi.ts";
import type { OfferType } from "@/features/offers/model/types.ts";
import { getStatusCode } from "@/shared/utils/getStatusCode.ts";
import ReviewEditorDialog from "@/widgets/reviews/ReviewEditorDialog.tsx";

interface PendingReviewCardProps {
  review: PendingDealReview;
  dealName?: string;
}

const subjectTypeLabel: Record<OfferType, string> = {
  good: "Товар",
  service: "Услуга",
};

function getCreateErrorMessage(error: FetchBaseQueryError | SerializedError | undefined): string | null {
  if (!error) {
    return null;
  }

  const code = getStatusCode(error);
  switch (code) {
    case 400:
      return "Отзыв сейчас нельзя создать для этой позиции";
    case 403:
      return "У вас нет прав на создание этого отзыва";
    case 404:
      return "Сделка или позиция не найдены";
    case 409:
      return "Отзыв по этому контексту уже существует";
    default:
      return "Не удалось сохранить отзыв";
  }
}

function PendingReviewCard({ review, dealName }: PendingReviewCardProps) {
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [createReview, { isLoading, error }] = reviewsApi.useCreateDealItemReviewMutation();
  const { data: currentUser } = usersApi.useGetCurrentUserQuery();
  const { data: provider } = usersApi.useGetUserByIdQuery(review.providerId ?? "", {
    skip: !review.providerId,
  });
  const { data: deal } = dealsApi.useGetDealByIdQuery(review.dealId, {
    skip: !review.itemRef && !review.offerRef,
  });
  const { data: offer } = offersApi.useGetOfferByIdQuery(review.offerRef?.offerId ?? "", {
    skip: !review.offerRef,
  });

  const itemRef = useMemo(() => {
    if (review.itemRef) {
      return review.itemRef;
    }

    if (!deal || !review.offerRef) {
      return null;
    }

    const resolvedItem = deal.items.find((item) =>
      item.offerId === review.offerRef?.offerId &&
      item.providerId === review.providerId &&
      item.receiverId === currentUser?.id,
    );

    return resolvedItem
      ? {
          dealId: review.dealId,
          itemId: resolvedItem.id,
        }
      : null;
  }, [currentUser?.id, deal, review.dealId, review.itemRef, review.offerRef, review.providerId]);
  const contextLabel = useMemo(() => reviewContextLabels[review.contextType], [review.contextType]);
  const subject = useMemo(() => {
    if (itemRef && deal) {
      const item = deal.items.find((dealItem) => dealItem.id === itemRef.itemId);
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
  }, [deal, itemRef, offer]);

  const handleCreate = async ({ rating, comment }: { rating: number; comment: string }) => {
    if (!itemRef) {
      return;
    }

    await createReview({
      dealId: itemRef.dealId,
      itemId: itemRef.itemId,
      body: {
        rating,
        ...(comment.trim() ? { comment: comment.trim() } : {}),
      },
    }).unwrap();

    setIsDialogOpen(false);
  };

  return (
    <>
      <Card variant="outlined">
        <CardContent>
          <Stack spacing={1.25}>
            <Box display="flex" justifyContent="space-between" gap={2} flexWrap="wrap">
              <div>
                <Typography variant="subtitle1" fontWeight={700}>
                  {dealName || "Завершенная сделка"}
                </Typography>
                <Typography variant="body2" color="text.secondary">
                  Поставщик: {provider?.name?.trim() || "Имя не указано"}
                </Typography>
                {subject && (
                  <Typography variant="body2" fontWeight={600} mt={0.5}>
                    {subject.label}: {subject.name}
                  </Typography>
                )}
              </div>
              <Chip label={contextLabel} size="small" color="success" variant="outlined" />
            </Box>

            <Box display="flex" gap={1} flexWrap="wrap">
              <Button component={RouterLink} to={`/deals/${review.dealId}`} size="small">
                Открыть сделку
              </Button>
              {itemRef && (
                <Button
                  component={RouterLink}
                  to={`/deals/${itemRef.dealId}/items/${itemRef.itemId}`}
                  size="small"
                >
                  Позиция
                </Button>
              )}
              {review.offerRef && (
                <Button component={RouterLink} to={`/offers/${review.offerRef.offerId}`} size="small">
                  Offer
                </Button>
              )}
              {review.providerId && (
                <Button component={RouterLink} to={`/users/${review.providerId}/reviews`} size="small">
                  Отзывы о поставщике
                </Button>
              )}
            </Box>

            <Box>
              <Button variant="contained" onClick={() => setIsDialogOpen(true)} disabled={!itemRef}>
                Оставить отзыв
              </Button>
            </Box>

            {!itemRef && (
              <Alert severity="warning">
                Не удалось определить позицию сделки для этого отзыва. Откройте сделку и попробуйте позже.
              </Alert>
            )}

            {error && <Alert severity="error">{getCreateErrorMessage(error)}</Alert>}
          </Stack>
        </CardContent>
      </Card>

      <ReviewEditorDialog
        open={isDialogOpen}
        title="Новый отзыв"
        submitLabel="Опубликовать"
        isLoading={isLoading}
        errorMessage={getCreateErrorMessage(error)}
        onClose={() => setIsDialogOpen(false)}
        onSubmit={handleCreate}
      />
    </>
  );
}

export default PendingReviewCard;
