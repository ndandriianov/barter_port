import { useEffect, useRef, useState } from "react";
import { Link as RouterLink, useNavigate, useParams } from "react-router-dom";
import { Alert, Box, Button, CircularProgress, Dialog, DialogContent, Divider, ImageList, ImageListItem, Typography } from "@mui/material";
import offersApi from "@/features/offers/api/offersApi";
import usersApi from "@/features/users/api/usersApi";
import useDraftOfferCounts from "@/features/deals/model/useDraftOfferCounts.ts";
import reviewsApi from "@/features/reviews/api/reviewsApi.ts";
import OfferCard from "@/widgets/offers/OfferCard";
import RespondToOfferModal from "@/widgets/offers/RespondToOfferModal";
import ReviewSummaryCard from "@/widgets/reviews/ReviewSummaryCard.tsx";

function OfferPage() {
  const { offerId } = useParams<{ offerId: string }>();
  const navigate = useNavigate();
  const [isRespondModalOpen, setIsRespondModalOpen] = useState(false);
  const [openedPhotoUrl, setOpenedPhotoUrl] = useState<string | null>(null);
  const viewedOfferIdsRef = useRef<Set<string>>(new Set());
  const { data: meData } = usersApi.useGetCurrentUserQuery();
  const [deleteOffer, { isLoading: isDeleting, error: deleteError }] = offersApi.useDeleteOfferMutation();
  const [viewOfferById] = offersApi.useViewOfferByIdMutation();

  const { data: offer, isLoading, error } = offersApi.useGetOfferByIdQuery(offerId ?? "", {
    skip: !offerId,
  });
  const { data: reviewsSummary } = reviewsApi.useGetOfferReviewsSummaryQuery(offerId ?? "", {
    skip: !offerId,
  });
  const isOwnOffer = !!meData && !!offer && offer.authorId === meData.id;
  const { countsByOfferId } = useDraftOfferCounts({ enabled: isOwnOffer });

  useEffect(() => {
    if (!offer || !meData || offer.authorId === meData.id || viewedOfferIdsRef.current.has(offer.id)) {
      return;
    }

    viewedOfferIdsRef.current.add(offer.id);
    void viewOfferById(offer.id).unwrap().catch(() => {
      viewedOfferIdsRef.current.delete(offer.id);
    });
  }, [meData, offer, viewOfferById]);

  if (!offerId) return <Alert severity="warning">Объявление не найдено</Alert>;

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (error || !offer) {
    return <Alert severity="warning">Объявление не найдено</Alert>;
  }

  const canRespond = !!meData && offer.authorId !== meData.id;

  const handleDelete = async () => {
    if (!window.confirm("Удалить объявление?")) {
      return;
    }

    try {
      await deleteOffer(offer.id).unwrap();
      navigate("/offers?tab=mine", { replace: true });
    } catch {
      // The error is surfaced via RTK Query state.
    }
  };

  return (
    <Box maxWidth={700} mx="auto">
      <Button
        size="small"
        variant="text"
        onClick={() => window.history.length > 1 ? navigate(-1) : navigate("/offers")}
        sx={{ mb: 2 }}
      >
        ← Назад
      </Button>

      <Typography variant="h4" fontWeight={700} mb={3}>
        {offer.name}
      </Typography>

      <OfferCard
        offer={offer}
        draftCount={isOwnOffer ? (countsByOfferId[offer.id] ?? 0) : 0}
        draftsHref={
          isOwnOffer && (countsByOfferId[offer.id] ?? 0) > 0
            ? `/deals/drafts?offerId=${offer.id}`
            : undefined
        }
        onPhotoClick={setOpenedPhotoUrl}
      />

      {offer.photoUrls.length > 1 && (
        <Box mt={3}>
          <Typography variant="h6" fontWeight={600} mb={1.5}>
            Ещё фото
          </Typography>
          <ImageList cols={2} gap={12} sx={{ m: 0 }}>
            {offer.photoUrls.slice(1).map((photoUrl) => (
              <ImageListItem key={photoUrl} sx={{ borderRadius: 2, overflow: "hidden" }}>
                <Box
                  component="img"
                  src={photoUrl}
                  alt={offer.name}
                  onClick={() => setOpenedPhotoUrl(photoUrl)}
                  sx={{
                    width: "100%",
                    height: 240,
                    objectFit: "cover",
                    display: "block",
                    cursor: "zoom-in",
                  }}
                />
              </ImageListItem>
            ))}
          </ImageList>
        </Box>
      )}

      <Divider sx={{ my: 3 }} />

      {reviewsSummary && (
        <Box mb={3}>
          <ReviewSummaryCard title="Отзывы по этому offer" summary={reviewsSummary} />
        </Box>
      )}

      {deleteError && (
        <Alert severity="error" sx={{ mb: 3 }}>
          Не удалось удалить объявление
        </Alert>
      )}

      <Box display="flex" gap={2} flexWrap="wrap">
        {canRespond && (
          <Button variant="contained" onClick={() => setIsRespondModalOpen(true)}>
            Откликнуться
          </Button>
        )}
        {isOwnOffer && (
          <Button component={RouterLink} to={`/offers/${offer.id}/edit`} variant="contained">
            Редактировать
          </Button>
        )}
        {isOwnOffer && (
          <Button variant="outlined" color="error" onClick={handleDelete} disabled={isDeleting}>
            {isDeleting ? "Удаление..." : "Удалить"}
          </Button>
        )}
        <Button component={RouterLink} to={`/offers/${offer.id}/reviews`} variant="outlined">
          Смотреть отзывы
        </Button>
        <Button component={RouterLink} to={`/users/${offer.authorId}/reviews`} variant="outlined">
          Отзывы о поставщике
        </Button>
        <Button variant="outlined" color="error">
          Пожаловаться
        </Button>
      </Box>

      <RespondToOfferModal
        targetOffer={offer}
        isOpen={isRespondModalOpen}
        onClose={() => setIsRespondModalOpen(false)}
      />

      <Dialog
        open={openedPhotoUrl !== null}
        onClose={() => setOpenedPhotoUrl(null)}
        maxWidth="lg"
        fullWidth
      >
        <DialogContent sx={{ p: 1.5, bgcolor: "common.black" }}>
          {openedPhotoUrl && (
            <Box
              component="img"
              src={openedPhotoUrl}
              alt={offer.name}
              sx={{
                width: "100%",
                maxHeight: "85vh",
                objectFit: "contain",
                display: "block",
              }}
            />
          )}
        </DialogContent>
      </Dialog>
    </Box>
  );
}

export default OfferPage;
