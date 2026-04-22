import { useState } from "react";
import { Box, Button, Card, CardContent, CardMedia, Chip, IconButton, Tooltip, Typography } from "@mui/material";
import VisibilityOutlinedIcon from "@mui/icons-material/VisibilityOutlined";
import CalendarTodayOutlinedIcon from "@mui/icons-material/CalendarTodayOutlined";
import FavoriteBorderOutlinedIcon from "@mui/icons-material/FavoriteBorderOutlined";
import FavoriteRoundedIcon from "@mui/icons-material/FavoriteRounded";
import StarRoundedIcon from "@mui/icons-material/StarRounded";
import { Link as RouterLink } from "react-router-dom";
import type { Offer, OfferAction, OfferType } from "@/features/offers/model/types";
import reviewsApi from "@/features/reviews/api/reviewsApi.ts";
import usersApi from "@/features/users/api/usersApi.ts";
import offersApi from "@/features/offers/api/offersApi.ts";
import {
  getOfferModerationLabel,
  getOfferModerationState,
} from "@/features/offers/model/getOfferModerationState.ts";
import UserAvatarLabel from "@/shared/UserAvatarLabel.tsx";

const actionLabels: Record<OfferAction, string> = {
  give: "Отдаю",
  take: "Ищу",
};

const typeLabels: Record<OfferType, string> = {
  good: "Товар",
  service: "Услуга",
};

const actionColors: Record<OfferAction, "success" | "primary"> = {
  give: "success",
  take: "primary",
};

const formatCreatedAt = (value: string) =>
  new Intl.DateTimeFormat("ru-RU", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
  }).format(new Date(value));

interface OfferCardProps {
  offer: Offer;
  showRating?: boolean;
  draftCount?: number;
  offerHref?: string;
  draftsHref?: string;
  onPhotoClick?: (photoUrl: string) => void;
  onFavoriteChange?: (offerId: string, isFavorite: boolean) => void;
  showModerationState?: boolean;
  isMine?: boolean;
}

function OfferCard({
  offer,
  showRating = false,
  draftCount = 0,
  offerHref,
  draftsHref,
  onPhotoClick,
  onFavoriteChange,
  showModerationState = false,
  isMine = false,
}: OfferCardProps) {
  const authorName = offer.authorName?.trim() || "Имя не указано";
  const { data: author } = usersApi.useGetUserByIdQuery(offer.authorId);
  const { data: summary } = reviewsApi.useGetOfferReviewsSummaryQuery(offer.id, {
    skip: !showRating,
  });
  const { data: offerReports } = offersApi.useGetOfferReportsQuery(offer.id, {
    skip: !showModerationState,
  });
  const [addOfferToFavorites, { isLoading: isAddingToFavorites }] = offersApi.useAddOfferToFavoritesMutation();
  const [removeOfferFromFavorites, { isLoading: isRemovingFromFavorites }] = offersApi.useRemoveOfferFromFavoritesMutation();
  const moderationState = getOfferModerationState(offer, offerReports);
  const moderationLabel = getOfferModerationLabel(moderationState);
  const isFavoriteActionLoading = isAddingToFavorites || isRemovingFromFavorites;
  const [favoriteOverride, setFavoriteOverride] = useState<boolean | null>(null);
  const isFavorite = favoriteOverride ?? offer.isFavorite;

  const handleToggleFavorite = async () => {
    const nextIsFavorite = !isFavorite;
    setFavoriteOverride(nextIsFavorite);
    onFavoriteChange?.(offer.id, nextIsFavorite);

    try {
      if (nextIsFavorite) {
        await addOfferToFavorites(offer.id).unwrap();
        return;
      }

      await removeOfferFromFavorites(offer.id).unwrap();
    } catch {
      setFavoriteOverride(null);
      onFavoriteChange?.(offer.id, !nextIsFavorite);
    }
  };

  return (
    <Card variant="outlined" sx={{ height: "100%", display: "flex", flexDirection: "column" }}>
      {offer.photoUrls.length > 0 && (
        <CardMedia
          component="img"
          image={offer.photoUrls[0]}
          alt={offer.name}
          onClick={onPhotoClick ? () => onPhotoClick(offer.photoUrls[0]) : undefined}
          sx={{
            height: 220,
            objectFit: "cover",
            borderBottom: 1,
            borderColor: "divider",
            cursor: onPhotoClick ? "zoom-in" : "default",
          }}
        />
      )}
      <CardContent sx={{ flexGrow: 1 }}>
        <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={1} mb={1}>
          <Box display="flex" gap={1} flexWrap="wrap">
            {isMine && <Chip label="Мой" size="small" color="secondary" />}
            <Chip label={typeLabels[offer.type]} size="small" variant="outlined" />
            <Chip label={actionLabels[offer.action]} size="small" color={actionColors[offer.action]} />
            {moderationLabel && (
              <Chip
                label={moderationLabel}
                size="small"
                color={moderationState === "hidden" ? "error" : moderationState === "pending" ? "warning" : "info"}
                variant={moderationState === "hidden" ? "filled" : "outlined"}
              />
            )}
            {draftCount > 0 && (
              <Chip label={`Черновики: ${draftCount}`} size="small" color="warning" variant="outlined" />
            )}
          </Box>
          <Tooltip title={isFavorite ? "Убрать из избранного" : "Добавить в избранное"}>
            <span>
              <IconButton
                size="small"
                color={isFavorite ? "error" : "default"}
                onClick={() => void handleToggleFavorite()}
                disabled={isFavoriteActionLoading}
              >
                {isFavorite ? <FavoriteRoundedIcon /> : <FavoriteBorderOutlinedIcon />}
              </IconButton>
            </span>
          </Tooltip>
        </Box>

        <Typography variant="h6" fontWeight={600} gutterBottom noWrap>
          {offer.name}
        </Typography>

        <Box sx={{ mb: 1 }}>
          <Typography variant="caption" color="text.secondary" display="block" sx={{ mb: 0.5 }}>
            Автор
          </Typography>
          <UserAvatarLabel
            userId={offer.authorId}
            name={author?.name ?? authorName}
            avatarUrl={author?.avatarUrl}
            size={30}
            textVariant="body2"
            fontWeight={400}
          />
        </Box>

        <Typography variant="body2" color="text.secondary" sx={{ mb: 2, display: "-webkit-box", WebkitLineClamp: 3, WebkitBoxOrient: "vertical", overflow: "hidden" }}>
          {offer.description}
        </Typography>

        {offer.tags.length > 0 && (
          <Box display="flex" gap={0.75} flexWrap="wrap" mb={2}>
            {offer.tags.map((tag) => (
              <Chip key={tag} label={`#${tag}`} size="small" variant="outlined" />
            ))}
          </Box>
        )}

        {showRating && (
          <Box display="flex" alignItems="center" gap={0.75} mb={2} color={summary && summary.count > 0 ? "warning.main" : "text.disabled"}>
            <StarRoundedIcon fontSize="small" />
            <Typography variant="body2" fontWeight={600} color="text.primary">
              {summary && summary.count > 0 ? summary.avgRating.toFixed(1) : "0.0"}
            </Typography>
            <Typography variant="caption" color="text.secondary">
              {summary && summary.count > 0 ? `(${summary.count})` : "(нет отзывов)"}
            </Typography>
          </Box>
        )}

        <Box display="flex" justifyContent="space-between" alignItems="center" mt="auto">
          <Box display="flex" alignItems="center" gap={0.5} color="text.disabled">
            <VisibilityOutlinedIcon fontSize="small" />
            <Typography variant="caption">{offer.views}</Typography>
          </Box>
          <Box display="flex" alignItems="center" gap={0.5} color="text.disabled">
            <CalendarTodayOutlinedIcon fontSize="small" />
            <Typography variant="caption">{formatCreatedAt(offer.createdAt)}</Typography>
          </Box>
        </Box>

        {(offerHref || (draftsHref && draftCount > 0)) && (
          <Box display="flex" gap={1} flexWrap="wrap" mt={2}>
            {offerHref && (
              <Button component={RouterLink} to={offerHref} size="small" variant="outlined">
                Открыть
              </Button>
            )}
            {draftsHref && draftCount > 0 && (
              <Button component={RouterLink} to={draftsHref} size="small" variant="outlined" color="warning">
                Черновики: {draftCount}
              </Button>
            )}
          </Box>
        )}
      </CardContent>
    </Card>
  );
}

export default OfferCard;
