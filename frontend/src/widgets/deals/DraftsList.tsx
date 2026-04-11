import {useEffect, useMemo} from "react";
import { Link as RouterLink } from "react-router-dom";
import {
  Alert,
  Box,
  Chip,
  CircularProgress,
  FormControl,
  IconButton,
  InputLabel,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
  MenuItem,
  Select,
  Tooltip,
  Typography,
} from "@mui/material";
import RefreshIcon from "@mui/icons-material/Refresh";
import dealsApi from "@/features/deals/api/dealsApi";
import offersApi from "@/features/offers/api/offersApi.ts";
import usersApi from "@/features/users/api/usersApi.ts";
import {useAppDispatch, useAppSelector} from "@/hooks/redux.ts";
import type {Draft} from "@/features/deals/model/types.ts";
import type {User} from "@/features/users/model/types.ts";

interface DraftsListProps {
  mode: "all" | "mine" | "others";
  selectedOfferId: string;
  onSelectedOfferIdChange: (offerId: string) => void;
}

function DraftsList({ mode, selectedOfferId, onSelectedOfferIdChange }: DraftsListProps) {
  const dispatch = useAppDispatch();
  const { data: currentUser } = usersApi.useGetCurrentUserQuery();
  const {
    data: mineData,
    isLoading: isMineLoading,
    error: mineError,
    refetch: refetchMine,
    isFetching: isMineFetching,
  } = dealsApi.useGetMyDraftDealsQuery({
    createdByMe: true,
    participating: true,
  });
  const {
    data: othersData,
    isLoading: isOthersLoading,
    error: othersError,
    refetch: refetchOthers,
    isFetching: isOthersFetching,
  } = dealsApi.useGetMyDraftDealsQuery({
    createdByMe: false,
    participating: true,
  });
  const {
    data: myOffersData,
    isLoading: isOffersLoading,
    error: offersError,
  } = offersApi.useGetOffersQuery({
    sort: "ByTime",
    my: true,
    cursor_limit: 100,
  });

  const mineDrafts = mineData ?? [];
  const incomingRawDrafts = othersData ?? [];

  const draftIdsFromMine = useMemo(
    () => new Set(mineDrafts.map((draft) => draft.id)),
    [mineDrafts],
  );

  const baseDrafts = useMemo(() => {
    if (mode === "mine") {
      return mineDrafts;
    }

    if (mode === "others") {
      return incomingRawDrafts.filter((draft) => !draftIdsFromMine.has(draft.id));
    }

    const merged = new Map<string, typeof mineDrafts[number]>();

    mineDrafts.forEach((draft) => {
      merged.set(draft.id, draft);
    });
    incomingRawDrafts.forEach((draft) => {
      if (!merged.has(draft.id)) {
        merged.set(draft.id, draft);
      }
    });

    return Array.from(merged.values());
  }, [draftIdsFromMine, incomingRawDrafts, mineDrafts, mode]);

  const draftIds = useMemo(
    () => baseDrafts.map((draft) => draft.id),
    [baseDrafts],
  );

  const participantIds = useMemo(
    () => [...new Set(baseDrafts.flatMap((draft) => draft.participants))],
    [baseDrafts],
  );

  useEffect(() => {
    if (participantIds.length === 0) {
      return;
    }

    const subscriptions = participantIds.map((id) =>
      dispatch(usersApi.endpoints.getUserById.initiate(id)),
    );

    return () => {
      subscriptions.forEach((subscription) => subscription.unsubscribe());
    };
  }, [dispatch, participantIds]);

  useEffect(() => {
    if (draftIds.length === 0) {
      return;
    }

    const subscriptions = draftIds.map((draftId) =>
      dispatch(dealsApi.endpoints.getDraftDealById.initiate(draftId)),
    );

    return () => {
      subscriptions.forEach((subscription) => subscription.unsubscribe());
    };
  }, [dispatch, draftIds]);

  const usersById = useAppSelector((state) =>
    participantIds.reduce<Record<string, User | undefined>>((acc, id) => {
      acc[id] = usersApi.endpoints.getUserById.select(id)(state).data;
      return acc;
    }, {}),
  );
  const draftQueryStatesById = useAppSelector((state) =>
    draftIds.reduce<Record<string, { isLoading: boolean; isUninitialized: boolean }>>((acc, draftId) => {
      const query = dealsApi.endpoints.getDraftDealById.select(draftId)(state);
      acc[draftId] = {
        isLoading: query.isLoading,
        isUninitialized: query.isUninitialized,
      };
      return acc;
    }, {}),
  );

  const draftDetailsById = useAppSelector((state) =>
    draftIds.reduce<Record<string, Draft | undefined>>((acc, draftId) => {
      acc[draftId] = dealsApi.endpoints.getDraftDealById.select(draftId)(state).data;
      return acc;
    }, {}),
  );

  const getParticipantNames = (ids: string[]) =>
    ids.length === 0
      ? "имя не указано"
      : ids
          .map((id) => {
            const name = usersById[id]?.name?.trim();
            return name ? name : "имя не указано";
          })
          .join(", ");

  const getDraftDirection = (draftId: string): "incoming" | "outgoing" => {
    const details = draftDetailsById[draftId];
    if (details && currentUser) {
      return details.authorId === currentUser.id ? "outgoing" : "incoming";
    }

    if (draftIdsFromMine.has(draftId)) {
      return "outgoing";
    }

    return "incoming";
  };

  const filteredDrafts = useMemo(() => {
    if (baseDrafts.length === 0) {
      return [];
    }

    if (!selectedOfferId) {
      return baseDrafts;
    }

    return baseDrafts.filter((draft) =>
      draftDetailsById[draft.id]?.offers.some((offer) => offer.id === selectedOfferId),
    );
  }, [baseDrafts, draftDetailsById, selectedOfferId]);

  const isDraftDetailsLoading = selectedOfferId !== "" &&
    draftIds.some((draftId) => {
      const query = draftQueryStatesById[draftId];
      return !draftDetailsById[draftId] || query?.isLoading || query?.isUninitialized;
    });

  const isLoading = mode === "all"
    ? isMineLoading || isOthersLoading
    : mode === "mine"
      ? isMineLoading
      : isOthersLoading;

  const isFetching = mode === "all"
    ? isMineFetching || isOthersFetching
    : mode === "mine"
      ? isMineFetching
      : isOthersFetching;

  const error = mode === "all"
    ? mineError ?? othersError
    : mode === "mine"
      ? mineError
      : othersError;

  const handleRefetch = () => {
    if (mode !== "others") {
      void refetchMine();
    }

    if (mode !== "mine") {
      void refetchOthers();
    }
  };

  const emptyMessage = mode === "all"
    ? "У вас пока нет черновиков"
    : mode === "mine"
      ? "У вас пока нет своих черновиков"
      : "Пока нет чужих черновиков с вашим участием";

  if (isLoading) {
    return (
      <Box display="flex" justifyContent="center" py={6}>
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return <Alert severity="error">Не удалось загрузить черновики</Alert>;
  }

  if (mode === "all" && !mineData && !othersData) {
    return <Alert severity="info">Черновики недоступны</Alert>;
  }

  if (mode === "mine" && !mineData) {
    return <Alert severity="info">Черновики недоступны</Alert>;
  }

  if (mode === "others" && !othersData) {
    return <Alert severity="info">Черновики недоступны</Alert>;
  }

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" gap={2} mb={2} flexWrap="wrap">
        <FormControl size="small" sx={{ minWidth: 260 }}>
          <InputLabel>Фильтр по моему объявлению</InputLabel>
          <Select
            value={selectedOfferId}
            label="Фильтр по моему объявлению"
            onChange={(event) => onSelectedOfferIdChange(event.target.value)}
            disabled={isOffersLoading || Boolean(offersError)}
          >
            <MenuItem value="">Все черновики</MenuItem>
            {(myOffersData?.offers ?? []).map((offer) => (
              <MenuItem key={offer.id} value={offer.id}>
                {offer.name}
              </MenuItem>
            ))}
          </Select>
        </FormControl>

        <Tooltip title="Обновить">
          <span>
            <IconButton onClick={handleRefetch} disabled={isFetching} size="small">
              <RefreshIcon />
            </IconButton>
          </span>
        </Tooltip>
      </Box>

      {offersError && (
        <Alert severity="warning" sx={{ mb: 2 }}>
          Не удалось загрузить ваши объявления для фильтрации.
        </Alert>
      )}

      {isDraftDetailsLoading ? (
        <Box display="flex" justifyContent="center" py={6}>
          <CircularProgress />
        </Box>
      ) : baseDrafts.length === 0 ? (
        <Typography color="text.secondary" textAlign="center" py={4}>
          {emptyMessage}
        </Typography>
      ) : filteredDrafts.length === 0 ? (
        <Typography color="text.secondary" textAlign="center" py={4}>
          {selectedOfferId
            ? "Черновики с выбранным объявлением не найдены"
            : emptyMessage}
        </Typography>
      ) : (
        <List disablePadding>
          {filteredDrafts.map((draft) => (
            <ListItem key={draft.id} disablePadding divider>
              <ListItemButton component={RouterLink} to={`/deals/drafts/${draft.id}`}>
                <ListItemText
                  primary={draft.name ?? "—"}
                  secondary={(
                    <Box display="flex" alignItems="center" gap={1} flexWrap="wrap">
                      <Typography variant="body2" color="text.secondary">
                        {getParticipantNames(draft.participants)}
                      </Typography>
                      {mode === "all" && (
                        <Chip
                          label={getDraftDirection(draft.id) === "incoming" ? "Входящий" : "Исходящий"}
                          size="small"
                          color={getDraftDirection(draft.id) === "incoming" ? "warning" : "success"}
                          variant="outlined"
                        />
                      )}
                    </Box>
                  )}
                />
              </ListItemButton>
            </ListItem>
          ))}
        </List>
      )}
    </Box>
  );
}

export default DraftsList;
