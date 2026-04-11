import {useEffect, useMemo, useState} from "react";
import { Link as RouterLink } from "react-router-dom";
import {
  Alert,
  Box,
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

function DraftsList() {
  const dispatch = useAppDispatch();
  const [selectedOfferId, setSelectedOfferId] = useState("");
  const { data, isLoading, error, refetch, isFetching } = dealsApi.useGetMyDraftDealsQuery({
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

  const draftIds = useMemo(
    () => (data ?? []).map((draft) => draft.id),
    [data],
  );

  const participantIds = useMemo(
    () => [...new Set((data ?? []).flatMap((draft) => draft.participants))],
    [data],
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

  const filteredDrafts = useMemo(() => {
    if (!data) {
      return [];
    }

    if (!selectedOfferId) {
      return data;
    }

    return data.filter((draft) =>
      draftDetailsById[draft.id]?.offers.some((offer) => offer.id === selectedOfferId),
    );
  }, [data, draftDetailsById, selectedOfferId]);

  const isDraftDetailsLoading = selectedOfferId !== "" &&
    draftIds.some((draftId) => {
      const query = draftQueryStatesById[draftId];
      return !draftDetailsById[draftId] || query?.isLoading || query?.isUninitialized;
    });

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

  if (!data) {
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
            onChange={(event) => setSelectedOfferId(event.target.value)}
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
            <IconButton onClick={() => refetch()} disabled={isFetching} size="small">
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
      ) : data.length === 0 ? (
        <Typography color="text.secondary" textAlign="center" py={4}>
          У вас пока нет черновых договоров
        </Typography>
      ) : filteredDrafts.length === 0 ? (
        <Typography color="text.secondary" textAlign="center" py={4}>
          Черновики с выбранным объявлением не найдены
        </Typography>
      ) : (
        <List disablePadding>
          {filteredDrafts.map((draft) => (
            <ListItem key={draft.id} disablePadding divider>
              <ListItemButton component={RouterLink} to={`/deals/drafts/${draft.id}`}>
                <ListItemText
                  primary={draft.name ?? "—"}
                  secondary={getParticipantNames(draft.participants)}
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
