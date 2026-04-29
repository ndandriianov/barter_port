import { useEffect, useRef, useState } from "react";
import { Link as RouterLink, useNavigate, useParams } from "react-router-dom";
import { Alert, Box, Button, CircularProgress, Dialog, DialogContent, Divider, ImageList, ImageListItem, Typography } from "@mui/material";
import FavoriteBorderOutlinedIcon from "@mui/icons-material/FavoriteBorderOutlined";
import FavoriteRoundedIcon from "@mui/icons-material/FavoriteRounded";
import offersApi from "@/features/offers/api/offersApi";
import usersApi from "@/features/users/api/usersApi";
import reviewsApi from "@/features/reviews/api/reviewsApi.ts";
import YandexMapReadonly from "@/shared/ui/YandexMapReadonly";
import OfferCard from "@/widgets/offers/OfferCard";
import RespondToOfferModal from "@/widgets/offers/RespondToOfferModal";
import ReviewSummaryCard from "@/widgets/reviews/ReviewSummaryCard.tsx";
import CreateOfferReportDialog from "@/widgets/offers/CreateOfferReportDialog.tsx";
import { getStatusCode } from "@/shared/utils/getStatusCode.ts";

function OfferPage() {
  const { offerId } = useParams<{ offerId: string }>();
  const navigate = useNavigate();
  const [isRespondModalOpen, setIsRespondModalOpen] = useState(false);
  const [isReportDialogOpen, setIsReportDialogOpen] = useState(false);
  const [openedPhotoUrl, setOpenedPhotoUrl] = useState<string | null>(null);
  const [reportSuccessMessage, setReportSuccessMessage] = useState<string | null>(null);
  const [favoriteOverride, setFavoriteOverride] = useState<{ offerId: string; value: boolean } | null>(null);
  const viewedOfferIdsRef = useRef<Set<string>>(new Set());
  const { data: meData } = usersApi.useGetCurrentUserQuery();
  const [deleteOffer, { isLoading: isDeleting, error: deleteError }] = offersApi.useDeleteOfferMutation();
  const [hideOfferByAuthor, { isLoading: isHiding, error: hideError }] = offersApi.useHideOfferByAuthorMutation();
  const [unhideOfferByAuthor, { isLoading: isUnhiding, error: unhideError }] = offersApi.useUnhideOfferByAuthorMutation();
  const [viewOfferById] = offersApi.useViewOfferByIdMutation();
  const [addOfferToFavorites, { isLoading: isAddingToFavorites }] = offersApi.useAddOfferToFavoritesMutation();
  const [removeOfferFromFavorites, { isLoading: isRemovingFromFavorites }] = offersApi.useRemoveOfferFromFavoritesMutation();

  const { data: offer, isLoading, error } = offersApi.useGetOfferByIdQuery(offerId ?? "", {
    skip: !offerId,
  });
  const { data: reviewsSummary } = reviewsApi.useGetOfferReviewsSummaryQuery(offerId ?? "", {
    skip: !offerId,
  });
  const isAdmin = meData?.isAdmin === true;
  const isOwnOffer = !!meData && !!offer && offer.authorId === meData.id;
  const isFavoriteActionLoading = isAddingToFavorites || isRemovingFromFavorites;
  const isVisibilityActionLoading = isHiding || isUnhiding;

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
  const displayedIsFavorite = favoriteOverride?.offerId === offer.id ? favoriteOverride.value : offer.isFavorite;
  const offerLocation = offer.latitude != null && offer.longitude != null
    ? { lat: offer.latitude, lon: offer.longitude }
    : null;
  const displayedOffer = {
    ...offer,
    isFavorite: displayedIsFavorite,
  };

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

  const handleToggleHidden = async () => {
    try {
      if (offer.hiddenByAuthor) {
        await unhideOfferByAuthor(offer.id).unwrap();
        return;
      }

      await hideOfferByAuthor(offer.id).unwrap();
    } catch {
      // The error is surfaced via RTK Query state.
    }
  };

  const hideActionLabel = offer.hiddenByAuthor
    ? offer.isHidden
      ? "Снять авторское скрытие"
      : "Убрать из скрытых"
    : "Скрыть";

  const visibilityActionError = hideError ?? unhideError;

  const handleToggleFavorite = async () => {
    const nextIsFavorite = !displayedIsFavorite;
    setFavoriteOverride({ offerId: offer.id, value: nextIsFavorite });

    try {
      if (nextIsFavorite) {
        await addOfferToFavorites(offer.id).unwrap();
        return;
      }

      await removeOfferFromFavorites(offer.id).unwrap();
    } catch {
      setFavoriteOverride(null);
    }
  };

  const handleFavoriteChange = (_offerId: string, isFavorite: boolean) => {
    setFavoriteOverride({ offerId: offer.id, value: isFavorite });
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
        {displayedOffer.name}
      </Typography>

      <OfferCard
        offer={displayedOffer}
        isMine={isOwnOffer}
        onFavoriteChange={handleFavoriteChange}
        showModerationState={isOwnOffer || isAdmin}
        draftCount={isOwnOffer ? offer.draftsCount : 0}
        draftsHref={
          isOwnOffer && offer.draftsCount > 0
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

      {offerLocation && (
        <Box mt={3}>
          <Typography variant="h6" fontWeight={600} mb={1.5}>
            Местоположение
          </Typography>
          <YandexMapReadonly value={offerLocation} height="260px" />
          <Typography variant="caption" color="text.secondary" mt={0.5} display="block">
            {offerLocation.lat.toFixed(6)}, {offerLocation.lon.toFixed(6)}
          </Typography>
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
          {getStatusCode(deleteError) === 409
            ? "Не удалось удалить объявление: по нему идет разбирательство."
            : "Не удалось удалить объявление"}
        </Alert>
      )}

      {visibilityActionError && (
        <Alert severity="error" sx={{ mb: 3 }}>
          Не удалось изменить режим скрытия объявления
        </Alert>
      )}

      {reportSuccessMessage && (
        <Alert severity="success" sx={{ mb: 3 }} onClose={() => setReportSuccessMessage(null)}>
          {reportSuccessMessage}
        </Alert>
      )}

      {isOwnOffer && offer.hiddenByAuthor && (
        <Alert severity="warning" sx={{ mb: 3 }}>
          {offer.isHidden
            ? "Вы скрыли это объявление сами. Сейчас оно также скрыто модератором, поэтому после снятия авторского скрытия оно все равно останется недоступным для других пользователей."
            : "Вы скрыли это объявление. Оно не показывается другим пользователям в каталоге и по прямой ссылке, пока вы не уберете его из скрытых."}
        </Alert>
      )}

      {(isOwnOffer || isAdmin) && offer.isHidden && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {isOwnOffer
            ? "Объявление скрыто модератором. Оно остается в вашем списке, но недоступно для других пользователей."
            : "Объявление скрыто модератором и недоступно обычным пользователям."}
        </Alert>
      )}

      {isOwnOffer && !offer.isHidden && offer.modificationBlocked && (
        <Alert severity="warning" sx={{ mb: 3 }}>
          По объявлению идет разбирательство. Пока жалоба на модерации, редактирование и удаление
          недоступны.
        </Alert>
      )}

      <Box display="flex" gap={2} flexWrap="wrap">
        {canRespond && (
          <Button variant="contained" onClick={() => setIsRespondModalOpen(true)}>
            Откликнуться
          </Button>
        )}
        {canRespond && (
          <Button
            variant={displayedIsFavorite ? "contained" : "outlined"}
            color={displayedIsFavorite ? "error" : "inherit"}
            startIcon={displayedIsFavorite ? <FavoriteRoundedIcon /> : <FavoriteBorderOutlinedIcon />}
            onClick={() => void handleToggleFavorite()}
            disabled={isFavoriteActionLoading}
          >
            {displayedIsFavorite ? "В избранном" : "В избранное"}
          </Button>
        )}
        {isOwnOffer && (
          <Button
            variant={offer.hiddenByAuthor ? "contained" : "outlined"}
            color={offer.hiddenByAuthor ? "warning" : "inherit"}
            onClick={() => void handleToggleHidden()}
            disabled={isVisibilityActionLoading}
          >
            {isVisibilityActionLoading
              ? "Сохраняем..."
              : hideActionLabel}
          </Button>
        )}
        {isOwnOffer && (
          <Button
            component={RouterLink}
            to={`/offers/${offer.id}/edit`}
            variant="contained"
            disabled={offer.isHidden || offer.modificationBlocked}
          >
            Редактировать
          </Button>
        )}
        {isOwnOffer && (
          <Button
            variant="outlined"
            color="error"
            onClick={handleDelete}
            disabled={isDeleting || offer.isHidden || offer.modificationBlocked}
          >
            {isDeleting ? "Удаление..." : "Удалить"}
          </Button>
        )}
        <Button component={RouterLink} to={`/offers/${offer.id}/reviews`} variant="outlined">
          Смотреть отзывы
        </Button>
        <Button component={RouterLink} to={`/users/${offer.authorId}/reviews`} variant="outlined">
          Отзывы о поставщике
        </Button>
        <Button
          variant="outlined"
          color="error"
          onClick={() => setIsReportDialogOpen(true)}
          disabled={!canRespond}
        >
          Пожаловаться
        </Button>
      </Box>

      <RespondToOfferModal
        targetOffer={displayedOffer}
        isOpen={isRespondModalOpen}
        onClose={() => setIsRespondModalOpen(false)}
      />

      {offerId && (
        <CreateOfferReportDialog
          offerId={offerId}
          open={isReportDialogOpen}
          onClose={() => setIsReportDialogOpen(false)}
          onSuccess={() => setReportSuccessMessage("Жалоба отправлена и добавлена в очередь модерации.")}
        />
      )}

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
