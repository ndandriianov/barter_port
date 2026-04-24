import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import {
  Alert,
  Box,
  Checkbox,
  Chip,
  CircularProgress,
  FormControl,
  FormControlLabel,
  Grid,
  IconButton,
  InputLabel,
  MenuItem,
  Select,
  TextField,
  Tooltip,
  Typography,
} from "@mui/material";
import RefreshIcon from "@mui/icons-material/Refresh";
import offersApi from "@/features/offers/api/offersApi";
import usersApi from "@/features/users/api/usersApi.ts";
import type {
  FavoriteOffersCursor,
  FavoritedOffer,
  GetFavoriteOffersParams,
  GetOffersParams,
  GetSubscribedOffersParams,
  Offer,
  SortType,
  UniversalCursor,
} from "@/features/offers/model/types";
import { normalizeOfferTags, parseOfferTagsInput } from "@/features/offers/model/tagUtils.ts";
import useDraftOfferCounts from "@/features/deals/model/useDraftOfferCounts.ts";
import OfferCard from "@/widgets/offers/OfferCard";

interface OffersListWidgetProps {
  mode: "mine" | "others" | "subscriptions" | "favorites";
}

const PAGE_SIZE = 8;

const mergeOffers = (currentOffers: Offer[], nextOffers: Offer[]) => {
  const offersById = new Map(currentOffers.map((offer) => [offer.id, offer]));

  for (const offer of nextOffers) {
    offersById.set(offer.id, offer);
  }

  return Array.from(offersById.values());
};

const buildOffersParams = (
  sortType: SortType,
  isMyOffers: boolean,
  cursor: UniversalCursor | null,
  tags: string[],
  withoutTags: boolean,
  location: { lat: number; lon: number } | null,
): GetOffersParams => {
  const params: GetOffersParams = {
    sort: sortType,
    my: isMyOffers,
    cursor_limit: PAGE_SIZE,
  };

  if (withoutTags) {
    params.withoutTags = true;
  } else if (tags.length > 0) {
    params.tags = tags;
  }

  if (sortType === "ByDistance" && location) {
    params.user_lat = location.lat;
    params.user_lon = location.lon;
  }

  if (!cursor) {
    return params;
  }

  params.cursor_id = cursor.id;

  if (sortType === "ByDistance" && typeof cursor.distance === "number") {
    params.cursor_distance = cursor.distance;
  }

  if (sortType === "ByTime" && cursor.createdAt) {
    params.cursor_created_at = cursor.createdAt;
  }

  if (sortType === "ByPopularity" && typeof cursor.views === "number") {
    params.cursor_views = cursor.views;
  }

  return params;
};

const buildSubscribedOffersParams = (
  sortType: SortType,
  cursor: UniversalCursor | null,
  location: { lat: number; lon: number } | null,
): GetSubscribedOffersParams => {
  const params: GetSubscribedOffersParams = {
    sort: sortType,
    cursor_limit: PAGE_SIZE,
  };

  if (sortType === "ByDistance" && location) {
    params.user_lat = location.lat;
    params.user_lon = location.lon;
  }

  if (!cursor) {
    return params;
  }

  params.cursor_id = cursor.id;

  if (sortType === "ByDistance" && typeof cursor.distance === "number") {
    params.cursor_distance = cursor.distance;
  }

  if (sortType === "ByTime" && cursor.createdAt) {
    params.cursor_created_at = cursor.createdAt;
  }

  if (sortType === "ByPopularity" && typeof cursor.views === "number") {
    params.cursor_views = cursor.views;
  }

  return params;
};

const buildFavoriteOffersParams = (cursor: FavoriteOffersCursor | null): GetFavoriteOffersParams => {
  const params: GetFavoriteOffersParams = {
    cursor_limit: PAGE_SIZE,
  };

  if (!cursor) {
    return params;
  }

  params.cursor_id = cursor.id;
  params.cursor_favorited_at = cursor.favoritedAt;

  return params;
};

function OffersListWidget({ mode }: OffersListWidgetProps) {
  const [sortType, setSortType] = useState<SortType>("ByTime");
  const [offers, setOffers] = useState<Offer[]>([]);
  const [nextCursor, setNextCursor] = useState<UniversalCursor | null>(null);
  const [nextFavoriteCursor, setNextFavoriteCursor] = useState<FavoriteOffersCursor | null>(null);
  const [tagsInput, setTagsInput] = useState("");
  const [withoutTagsOnly, setWithoutTagsOnly] = useState(false);
  const [isInitialLoading, setIsInitialLoading] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [hasLoadedOnce, setHasLoadedOnce] = useState(false);
  const [initialError, setInitialError] = useState<string | null>(null);
  const [loadMoreError, setLoadMoreError] = useState<string | null>(null);
  const isMyOffers = mode === "mine";
  const isSubscribedOffers = mode === "subscriptions";
  const isFavoriteOffers = mode === "favorites";
  const sentinelRef = useRef<HTMLDivElement | null>(null);
  const nextCursorRef = useRef<UniversalCursor | null>(null);
  const nextFavoriteCursorRef = useRef<FavoriteOffersCursor | null>(null);
  const isInitialLoadingRef = useRef(true);
  const isLoadingMoreRef = useRef(false);
  const hasLoadedOnceRef = useRef(false);
  const feedKeyRef = useRef("");
  const [triggerGetOffers] = offersApi.useLazyGetOffersQuery();
  const [triggerGetSubscribedOffers] = offersApi.useLazyGetSubscribedOffersQuery();
  const [triggerGetFavoriteOffers] = offersApi.useLazyGetFavoriteOffersQuery();
  const { data: existingTags = [] } = offersApi.useListTagsQuery(undefined, {
    skip: isSubscribedOffers || isFavoriteOffers,
  });
  const { data: currentUser } = usersApi.useGetCurrentUserQuery();
  const { countsByOfferId } = useDraftOfferCounts({ enabled: isMyOffers });
  const currentLocation = useMemo(
    () => (
      currentUser?.currentLatitude != null && currentUser.currentLongitude != null
        ? { lat: currentUser.currentLatitude, lon: currentUser.currentLongitude }
        : null
    ),
    [currentUser?.currentLatitude, currentUser?.currentLongitude],
  );
  const parsedTags = useMemo(() => parseOfferTagsInput(tagsInput), [tagsInput]);
  const suggestedTags = useMemo(
    () => existingTags.filter((tag) => !parsedTags.includes(tag)).slice(0, 10),
    [existingTags, parsedTags],
  );
  const parsedTagsKey = parsedTags.join("|");
  const locationKey = currentLocation ? `${currentLocation.lat}:${currentLocation.lon}` : "none";
  const feedKey = `${mode}:${sortType}:${withoutTagsOnly}:${parsedTagsKey}:${locationKey}`;

  feedKeyRef.current = feedKey;
  nextCursorRef.current = nextCursor;
  nextFavoriteCursorRef.current = nextFavoriteCursor;

  useEffect(() => {
    if (sortType === "ByDistance" && !currentLocation) {
      setSortType("ByTime");
    }
  }, [currentLocation, sortType]);

  const loadOffersPage = useCallback(async (
    cursor: UniversalCursor | null,
    favoriteCursor: FavoriteOffersCursor | null,
    replace: boolean,
  ) => {
    const requestFeedKey = feedKey;

    if (replace) {
      if (!hasLoadedOnceRef.current) {
        setIsInitialLoading(true);
      }
      isInitialLoadingRef.current = true;
      setIsLoadingMore(false);
      isLoadingMoreRef.current = false;
      setInitialError(null);
      setLoadMoreError(null);
      setNextCursor(null);
      setNextFavoriteCursor(null);
      nextCursorRef.current = null;
      nextFavoriteCursorRef.current = null;
    } else {
      if ((!cursor && !favoriteCursor) || isInitialLoadingRef.current || isLoadingMoreRef.current) {
        return;
      }

      setLoadMoreError(null);
      setIsLoadingMore(true);
      isLoadingMoreRef.current = true;
    }

    if (!isFavoriteOffers && sortType === "ByDistance" && !currentLocation) {
      if (replace) {
        hasLoadedOnceRef.current = true;
        setHasLoadedOnce(true);
        setOffers([]);
        setInitialError("Чтобы сортировать по расстоянию, сначала укажите своё местоположение в профиле.");
      } else {
        setLoadMoreError("Не удалось загрузить следующие объявления без сохранённого местоположения.");
      }
      return;
    }

    try {
      if (isFavoriteOffers) {
        const response = await triggerGetFavoriteOffers(buildFavoriteOffersParams(favoriteCursor)).unwrap();

        if (feedKeyRef.current !== requestFeedKey) {
          return;
        }

        const nextOffers: Offer[] = response.offers as FavoritedOffer[];
        setOffers((currentOffers) => (replace ? nextOffers : mergeOffers(currentOffers, nextOffers)));
        setNextFavoriteCursor(response.nextCursor);
        nextFavoriteCursorRef.current = response.nextCursor;
        setNextCursor(null);
        nextCursorRef.current = null;
      } else {
        const response = isSubscribedOffers
          ? await triggerGetSubscribedOffers(buildSubscribedOffersParams(sortType, cursor, currentLocation)).unwrap()
          : await triggerGetOffers(buildOffersParams(sortType, isMyOffers, cursor, parsedTags, withoutTagsOnly, currentLocation)).unwrap();

        if (feedKeyRef.current !== requestFeedKey) {
          return;
        }

        setOffers((currentOffers) => (replace ? response.offers : mergeOffers(currentOffers, response.offers)));
        setNextCursor(response.nextCursor);
        nextCursorRef.current = response.nextCursor;
        setNextFavoriteCursor(null);
        nextFavoriteCursorRef.current = null;
      }

      if (replace) {
        hasLoadedOnceRef.current = true;
        setHasLoadedOnce(true);
      }
    } catch {
      if (feedKeyRef.current !== requestFeedKey) {
        return;
      }

      if (replace) {
        hasLoadedOnceRef.current = true;
        setHasLoadedOnce(true);
        setInitialError(
          isFavoriteOffers
            ? "Не удалось загрузить избранные объявления"
            : isSubscribedOffers
              ? "Не удалось загрузить объявления от подписок"
              : "Не удалось загрузить список объявлений",
        );
        setNextCursor(null);
        setNextFavoriteCursor(null);
        nextCursorRef.current = null;
        nextFavoriteCursorRef.current = null;
      } else {
        setLoadMoreError(
          isFavoriteOffers
            ? "Не удалось загрузить следующие избранные объявления"
            : isSubscribedOffers
              ? "Не удалось загрузить следующие объявления от подписок"
              : "Не удалось загрузить следующие объявления",
        );
      }
    } finally {
      if (replace) {
        if (feedKeyRef.current === requestFeedKey) {
          setIsInitialLoading(false);
          isInitialLoadingRef.current = false;
        }
      } else {
        isLoadingMoreRef.current = false;

        if (feedKeyRef.current === requestFeedKey) {
          setIsLoadingMore(false);
        }
      }
    }
  }, [
    feedKey,
    isFavoriteOffers,
    isMyOffers,
    isSubscribedOffers,
    parsedTags,
    sortType,
    currentLocation,
    triggerGetFavoriteOffers,
    triggerGetOffers,
    triggerGetSubscribedOffers,
    withoutTagsOnly,
  ]);

  useEffect(() => {
    void loadOffersPage(null, null, true);
  }, [loadOffersPage]);

  const hasNextCursor = nextCursor !== null || nextFavoriteCursor !== null;

  useEffect(() => {
    if (offers.length === 0) {
      return;
    }

    const sentinelNode = sentinelRef.current;

    if (!sentinelNode) {
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0];
        const nextRegularCursor = nextCursorRef.current;
        const nextFavoritesCursor = nextFavoriteCursorRef.current;

        if (
          !entry?.isIntersecting ||
          (!nextRegularCursor && !nextFavoritesCursor) ||
          isInitialLoadingRef.current ||
          isLoadingMoreRef.current
        ) {
          return;
        }

        void loadOffersPage(nextRegularCursor, nextFavoritesCursor, false);
      },
      {
        rootMargin: "300px 0px",
      },
    );

    observer.observe(sentinelNode);

    return () => observer.disconnect();
  }, [loadOffersPage, offers.length, isInitialLoading, hasNextCursor]);

  if (isInitialLoading && !hasLoadedOnce) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (initialError && !hasLoadedOnce) {
    return <Alert severity="error">{initialError}</Alert>;
  }

  const handleAddSuggestedTag = (tag: string) => {
    setTagsInput(normalizeOfferTags([...parsedTags, tag]).join(", "));
    setWithoutTagsOnly(false);
  };

  const handleRemoveTag = (tagToRemove: string) => {
    setTagsInput(normalizeOfferTags(parsedTags.filter((tag) => tag !== tagToRemove)).join(", "));
  };

  const handleFavoriteChange = (offerId: string, isFavorite: boolean) => {
    setOffers((currentOffers) => {
      if (isFavoriteOffers && !isFavorite) {
        return currentOffers.filter((offer) => offer.id !== offerId);
      }

      return currentOffers.map((offer) => (
        offer.id === offerId
          ? { ...offer, isFavorite }
          : offer
      ));
    });
  };

  return (
    <Box>
      {initialError && hasLoadedOnce && (
        <Alert severity="error" sx={{ mb: 3 }}>
          {initialError}
        </Alert>
      )}

      <Box display="flex" alignItems="center" gap={2} mb={3} flexWrap="wrap">
        <FormControl size="small" sx={{ minWidth: 200 }}>
          <InputLabel>Сортировка</InputLabel>
          <Select
            value={sortType}
            label="Сортировка"
            onChange={(e) => setSortType(e.target.value as SortType)}
            disabled={isFavoriteOffers}
          >
            <MenuItem value="ByTime">Сначала новые</MenuItem>
            <MenuItem value="ByPopularity">По популярности</MenuItem>
            <MenuItem value="ByDistance" disabled={!currentLocation}>
              Сначала ближе
            </MenuItem>
          </Select>
        </FormControl>

        <Tooltip title="Обновить">
          <span>
            <IconButton onClick={() => void loadOffersPage(null, null, true)} disabled={isInitialLoading || isLoadingMore}>
              <RefreshIcon />
            </IconButton>
          </span>
        </Tooltip>
      </Box>

      {!isFavoriteOffers && !currentLocation && (
        <Alert severity="info" sx={{ mb: 3 }}>
          Сортировка по расстоянию доступна после сохранения местоположения в профиле.
        </Alert>
      )}

      {!isSubscribedOffers && !isFavoriteOffers && (
        <Box mb={3}>
          <TextField
            label="Фильтр по тегам"
            value={tagsInput}
            onChange={(e) => {
              setTagsInput(e.target.value);
              if (e.target.value.trim() !== "") {
                setWithoutTagsOnly(false);
              }
            }}
            placeholder="bike, repair"
            fullWidth
            helperText="Через запятую. Будут показаны объявления, содержащие все указанные теги."
          />
          <Box mt={1} display="flex" alignItems="center" gap={2} flexWrap="wrap">
            <FormControlLabel
              control={
                <Checkbox
                  checked={withoutTagsOnly}
                  onChange={(e) => {
                    setWithoutTagsOnly(e.target.checked);
                    if (e.target.checked) {
                      setTagsInput("");
                    }
                  }}
                />
              }
              label="Только без тегов"
            />
            {parsedTags.length > 0 && (
              <Box display="flex" gap={0.75} flexWrap="wrap">
                {parsedTags.map((tag) => (
                  <Chip
                    key={tag}
                    label={`#${tag}`}
                    size="small"
                    clickable
                    onClick={() => handleRemoveTag(tag)}
                  />
                ))}
              </Box>
            )}
          </Box>
          {suggestedTags.length > 0 && !withoutTagsOnly && (
            <Box mt={1.5} display="flex" gap={0.75} flexWrap="wrap">
              {suggestedTags.map((tag) => (
                <Chip
                  key={tag}
                  label={tag}
                  size="small"
                  variant="outlined"
                  clickable
                  onClick={() => handleAddSuggestedTag(tag)}
                />
              ))}
            </Box>
          )}
        </Box>
      )}

      {offers.length === 0 ? (
        <Typography color="text.secondary" textAlign="center" py={6}>
          {isMyOffers
            ? "У вас пока нет объявлений"
            : isFavoriteOffers
              ? "Вы пока ничего не добавили в избранное"
              : isSubscribedOffers
              ? "Пока нет объявлений от ваших подписок"
              : "Пока нет объявлений"}
        </Typography>
      ) : (
        <>
          <Grid container spacing={2}>
            {offers.map((offer) => (
              <Grid key={offer.id} size={{ xs: 12, sm: 6, md: 4, lg: 3 }}>
                <OfferCard
                  offer={offer}
                  isMine={offer.authorId === currentUser?.id}
                  onFavoriteChange={handleFavoriteChange}
                  showRating
                  showModerationState={isMyOffers || currentUser?.isAdmin === true}
                  draftCount={isMyOffers ? (countsByOfferId[offer.id] ?? 0) : 0}
                  offerHref={`/offers/${offer.id}`}
                  draftsHref={
                    isMyOffers && (countsByOfferId[offer.id] ?? 0) > 0
                      ? `/deals/drafts?offerId=${offer.id}`
                      : undefined
                  }
                />
              </Grid>
            ))}
          </Grid>

          {(nextCursor || nextFavoriteCursor) && <Box ref={sentinelRef} sx={{ height: 1 }} />}

          <Box py={3}>
            {isLoadingMore ? (
              <Box display="flex" justifyContent="center">
                <CircularProgress size={28} />
              </Box>
            ) : loadMoreError ? (
              <Alert severity="error">{loadMoreError}</Alert>
            ) : !nextCursor && !nextFavoriteCursor ? (
              <Alert severity="warning" sx={{ justifyContent: "center" }}>
                Больше объявлений нет
              </Alert>
            ) : null}
          </Box>
        </>
      )}
    </Box>
  );
}

export default OffersListWidget;
