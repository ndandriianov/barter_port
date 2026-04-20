import { useEffect, useEffectEvent, useRef, useState } from "react";
import {
  Alert,
  Box,
  CircularProgress,
  FormControl,
  Grid,
  IconButton,
  InputLabel,
  MenuItem,
  Select,
  Tooltip,
  Typography,
} from "@mui/material";
import RefreshIcon from "@mui/icons-material/Refresh";
import offersApi from "@/features/offers/api/offersApi";
import usersApi from "@/features/users/api/usersApi.ts";
import type { GetOffersParams, Offer, SortType, UniversalCursor } from "@/features/offers/model/types";
import useDraftOfferCounts from "@/features/deals/model/useDraftOfferCounts.ts";
import OfferCard from "@/widgets/offers/OfferCard";

interface OffersListWidgetProps {
  mode: "mine" | "others";
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
): GetOffersParams => {
  const params: GetOffersParams = {
    sort: sortType,
    my: isMyOffers,
    cursor_limit: PAGE_SIZE,
  };

  if (!cursor) {
    return params;
  }

  params.cursor_id = cursor.id;

  if (sortType === "ByTime" && cursor.createdAt) {
    params.cursor_created_at = cursor.createdAt;
  }

  if (sortType === "ByPopularity" && typeof cursor.views === "number") {
    params.cursor_views = cursor.views;
  }

  return params;
};

function OffersListWidget({ mode }: OffersListWidgetProps) {
  const [sortType, setSortType] = useState<SortType>("ByTime");
  const [offers, setOffers] = useState<Offer[]>([]);
  const [nextCursor, setNextCursor] = useState<UniversalCursor | null>(null);
  const [isInitialLoading, setIsInitialLoading] = useState(true);
  const [isLoadingMore, setIsLoadingMore] = useState(false);
  const [initialError, setInitialError] = useState<string | null>(null);
  const [loadMoreError, setLoadMoreError] = useState<string | null>(null);
  const isMyOffers = mode === "mine";
  const sentinelRef = useRef<HTMLDivElement | null>(null);
  const nextCursorRef = useRef<UniversalCursor | null>(null);
  const isInitialLoadingRef = useRef(true);
  const isLoadingMoreRef = useRef(false);
  const feedKeyRef = useRef("");
  const [triggerGetOffers] = offersApi.useLazyGetOffersQuery();
  const { data: currentUser } = usersApi.useGetCurrentUserQuery();
  const { countsByOfferId } = useDraftOfferCounts({ enabled: isMyOffers });
  const feedKey = `${mode}:${sortType}`;

  feedKeyRef.current = feedKey;
  nextCursorRef.current = nextCursor;

  const loadOffersPage = useEffectEvent(async (cursor: UniversalCursor | null, replace: boolean) => {
    const requestFeedKey = feedKeyRef.current;

    if (replace) {
      setIsInitialLoading(true);
      isInitialLoadingRef.current = true;
      setIsLoadingMore(false);
      isLoadingMoreRef.current = false;
      setInitialError(null);
      setLoadMoreError(null);
      setOffers([]);
      setNextCursor(null);
      nextCursorRef.current = null;
    } else {
      if (!cursor || isInitialLoadingRef.current || isLoadingMoreRef.current) {
        return;
      }

      setLoadMoreError(null);
      setIsLoadingMore(true);
      isLoadingMoreRef.current = true;
    }

    try {
      const response = await triggerGetOffers(buildOffersParams(sortType, isMyOffers, cursor)).unwrap();

      if (feedKeyRef.current !== requestFeedKey) {
        return;
      }

      setOffers((currentOffers) => (replace ? response.offers : mergeOffers(currentOffers, response.offers)));
      setNextCursor(response.nextCursor);
      nextCursorRef.current = response.nextCursor;
    } catch {
      if (feedKeyRef.current !== requestFeedKey) {
        return;
      }

      if (replace) {
        setInitialError("Не удалось загрузить список объявлений");
        setOffers([]);
        setNextCursor(null);
        nextCursorRef.current = null;
      } else {
        setLoadMoreError("Не удалось загрузить следующие объявления");
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
  });

  useEffect(() => {
    void loadOffersPage(null, true);
  }, [mode, sortType]);

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

        if (!entry?.isIntersecting || !nextCursorRef.current || isInitialLoadingRef.current || isLoadingMoreRef.current) {
          return;
        }

        void loadOffersPage(nextCursorRef.current, false);
      },
      {
        rootMargin: "300px 0px",
      },
    );

    observer.observe(sentinelNode);

    return () => observer.disconnect();
  }, [offers.length, mode, sortType]);

  if (isInitialLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (initialError) {
    return <Alert severity="error">{initialError}</Alert>;
  }

  return (
    <Box>
      <Box display="flex" alignItems="center" gap={2} mb={3} flexWrap="wrap">
        <FormControl size="small" sx={{ minWidth: 200 }}>
          <InputLabel>Сортировка</InputLabel>
          <Select
            value={sortType}
            label="Сортировка"
            onChange={(e) => setSortType(e.target.value as SortType)}
          >
            <MenuItem value="ByTime">Сначала новые</MenuItem>
            <MenuItem value="ByPopularity">По популярности</MenuItem>
          </Select>
        </FormControl>

        <Tooltip title="Обновить">
          <span>
            <IconButton onClick={() => void loadOffersPage(null, true)} disabled={isInitialLoading || isLoadingMore}>
              <RefreshIcon />
            </IconButton>
          </span>
        </Tooltip>
      </Box>

      {offers.length === 0 ? (
        <Typography color="text.secondary" textAlign="center" py={6}>
          {isMyOffers ? "У вас пока нет объявлений" : "Пока нет объявлений"}
        </Typography>
      ) : (
        <>
          <Grid container spacing={2}>
            {offers.map((offer) => (
              <Grid key={offer.id} size={{ xs: 12, sm: 6, md: 4, lg: 3 }}>
                <OfferCard
                  offer={offer}
                  isMine={offer.authorId === currentUser?.id}
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

          <Box ref={sentinelRef} sx={{ height: 1 }} />

          <Box py={3}>
            {isLoadingMore ? (
              <Box display="flex" justifyContent="center">
                <CircularProgress size={28} />
              </Box>
            ) : loadMoreError ? (
              <Alert severity="error">{loadMoreError}</Alert>
            ) : !nextCursor ? (
              <Typography color="text.secondary" textAlign="center">
                Больше объявлений нет
              </Typography>
            ) : null}
          </Box>
        </>
      )}
    </Box>
  );
}

export default OffersListWidget;
