import {useEffect, useMemo} from "react";
import { Link as RouterLink } from "react-router-dom";
import {
  Alert,
  Box,
  CircularProgress,
  IconButton,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
  Tooltip,
  Typography,
} from "@mui/material";
import RefreshIcon from "@mui/icons-material/Refresh";
import dealsApi from "@/features/deals/api/dealsApi";
import usersApi from "@/features/users/api/usersApi.ts";
import {useAppDispatch, useAppSelector} from "@/hooks/redux.ts";
import type {User} from "@/features/users/model/types.ts";

function DraftsList() {
  const dispatch = useAppDispatch();
  const { data, isLoading, error, refetch, isFetching } = dealsApi.useGetMyDraftDealsQuery({
    createdByMe: false,
    participating: true,
  });

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

  const usersById = useAppSelector((state) =>
    participantIds.reduce<Record<string, User | undefined>>((acc, id) => {
      acc[id] = usersApi.endpoints.getUserById.select(id)(state).data;
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
      <Box display="flex" justifyContent="flex-end" mb={1}>
        <Tooltip title="Обновить">
          <span>
            <IconButton onClick={() => refetch()} disabled={isFetching} size="small">
              <RefreshIcon />
            </IconButton>
          </span>
        </Tooltip>
      </Box>

      {data.length === 0 ? (
        <Typography color="text.secondary" textAlign="center" py={4}>
          У вас пока нет черновых договоров
        </Typography>
      ) : (
        <List disablePadding>
          {data.map((draft) => (
            <ListItem key={draft.id} disablePadding divider>
              <ListItemButton component={RouterLink} to={`/deals/drafts/${draft.id}`}>
                <ListItemText
                  primary={getParticipantNames(draft.participants)}
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
