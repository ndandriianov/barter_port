import { useEffect, useMemo } from "react";
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
import usersApi from "@/features/users/api/usersApi.ts";
import { useAppDispatch, useAppSelector } from "@/hooks/redux.ts";
import type { User } from "@/features/users/model/types.ts";
import type { DealStatus, GetDealsResponse } from "@/features/deals/model/types.ts";
import { appRoutes } from "@/shared/config/appRoutes.ts";

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

interface DealsListProps {
  deals: GetDealsResponse;
  isLoading: boolean;
  isFetching: boolean;
  hasError: boolean;
  onRefresh: () => void;
  emptyMessage: string;
  showStatusSections?: boolean;
}

function DealsList({
  deals,
  isLoading,
  isFetching,
  hasError,
  onRefresh,
  emptyMessage,
  showStatusSections = false,
}: DealsListProps) {
  const dispatch = useAppDispatch();

  const participantIds = useMemo(
    () => [...new Set(deals.flatMap((deal) => deal.participants))],
    [deals],
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

  const groupedDeals = useMemo(
    () =>
      dealStatusOrder
        .map((status) => ({
          status,
          deals: deals.filter((deal) => deal.status === status),
        }))
        .filter((group) => group.deals.length > 0),
    [deals],
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

  if (hasError) {
    return <Alert severity="error">Не удалось загрузить сделки</Alert>;
  }

  return (
    <Box>
      <Box display="flex" alignItems="center" gap={2} mb={2} flexWrap="wrap">
        <Tooltip title="Обновить">
          <span>
            <IconButton onClick={onRefresh} disabled={isFetching} size="small">
              <RefreshIcon />
            </IconButton>
          </span>
        </Tooltip>
      </Box>

      {deals.length === 0 ? (
        <Typography color="text.secondary" textAlign="center" py={4}>
          {emptyMessage}
        </Typography>
      ) : (
        <Box display="flex" flexDirection="column" gap={3}>
          {groupedDeals.map(({ status, deals: statusDeals }) => (
            <Box key={status}>
              {showStatusSections && (
                <Typography variant="subtitle1" mb={1}>
                  {dealStatusLabels[status]}
                </Typography>
              )}
              <List disablePadding>
                {statusDeals.map((deal) => (
                  <ListItem key={deal.id} disablePadding divider>
                    <ListItemButton component={RouterLink} to={appRoutes.deals.detail(deal.id)}>
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
