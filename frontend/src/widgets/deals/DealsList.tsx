import { useEffect, useMemo, useState } from "react";
import { Link as RouterLink } from "react-router-dom";
import {
  Alert,
  Box,
  Checkbox,
  CircularProgress,
  FormControlLabel,
  FormGroup,
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
import { useAppDispatch, useAppSelector } from "@/hooks/redux.ts";
import type { User } from "@/features/users/model/types.ts";
import type { DealStatus, GetDealsResponse } from "@/features/deals/model/types.ts";

type DealListItem = GetDealsResponse[number];

const dealStatusOrder: DealStatus[] = [
  "LookingForParticipants",
  "Discussion",
  "Confirmed",
  "Completed",
  "Cancelled",
  "Failed",
];

const dealStatusLabels: Record<DealStatus, string> = {
  LookingForParticipants: "В поиске участников",
  Discussion: "Обсуждение",
  Confirmed: "Подтверждены",
  Completed: "Завершены",
  Cancelled: "Отменены",
  Failed: "Не состоялись",
};

function DealsList() {
  const dispatch = useAppDispatch();
  const [myOnly, setMyOnly] = useState(false);
  const [openOnly, setOpenOnly] = useState(false);

  const { data, isLoading, isFetching, error, refetch } = dealsApi.useGetDealsQuery({
    my: myOnly || undefined,
    open: openOnly || undefined,
  });

  const participantIds = useMemo(
    () => [...new Set((data ?? []).flatMap((deal) => deal.participants))],
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

  const groupedDeals = useMemo(() => {
    const groups = dealStatusOrder.reduce<Record<DealStatus, DealListItem[]>>(
      (acc, status) => {
        acc[status] = [];
        return acc;
      },
      {
        LookingForParticipants: [],
        Discussion: [],
        Confirmed: [],
        Completed: [],
        Cancelled: [],
        Failed: [],
      },
    );

    (data ?? []).forEach((deal) => {
      groups[deal.status].push(deal);
    });

    return dealStatusOrder
      .map((status) => ({ status, deals: groups[status] }))
      .filter((group) => group.deals.length > 0);
  }, [data]);

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
    return <Alert severity="error">Не удалось загрузить сделки</Alert>;
  }

  if (!data) {
    return <Alert severity="info">Список сделок недоступен</Alert>;
  }

  return (
    <Box>
      <Box display="flex" alignItems="center" gap={2} mb={2} flexWrap="wrap">
        <FormGroup row>
          <FormControlLabel
            control={
              <Checkbox
                checked={myOnly}
                onChange={(e) => setMyOnly(e.target.checked)}
                size="small"
              />
            }
            label="Только мои"
          />
          <FormControlLabel
            control={
              <Checkbox
                checked={openOnly}
                onChange={(e) => setOpenOnly(e.target.checked)}
                size="small"
              />
            }
            label="Только открытые"
          />
        </FormGroup>

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
          Сделок пока нет
        </Typography>
      ) : (
        <Box display="flex" flexDirection="column" gap={3}>
          {groupedDeals.map(({ status, deals }) => (
            <Box key={status}>
              <Typography variant="subtitle1" mb={1}>
                {dealStatusLabels[status]}
              </Typography>
              <List disablePadding>
                {deals.map((deal) => (
                  <ListItem key={deal.id} disablePadding divider>
                    <ListItemButton component={RouterLink} to={`/deals/${deal.id}`}>
                      <ListItemText
                        primary={deal.name ?? "—"}
                        secondary={getParticipantNames(deal.participants)}
                      />
                    </ListItemButton>
                  </ListItem>
                ))}
              </List>
            </Box>
          ))}
        </Box>
      )}
    </Box>
  );
}

export default DealsList;
