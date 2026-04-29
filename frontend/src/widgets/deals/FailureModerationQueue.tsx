import { useEffect, useMemo, useState } from "react";
import { Link as RouterLink } from "react-router-dom";
import {
  Alert,
  Box,
  Button,
  Card,
  CardContent,
  Chip,
  CircularProgress,
  Stack,
  Typography,
} from "@mui/material";
import usersApi from "@/features/users/api/usersApi.ts";
import { useAppDispatch, useAppSelector } from "@/hooks/redux.ts";
import type { User } from "@/features/users/model/types.ts";
import { getStatusCode } from "@/shared/utils/getStatusCode";
import dealsApi from "@/features/deals/api/dealsApi";
import FailureResolutionDialog from "@/widgets/deals/FailureResolutionDialog.tsx";

function FailureModerationQueue() {
  const dispatch = useAppDispatch();
  const [selectedDealId, setSelectedDealId] = useState<string | null>(null);

  const { data: deals = [], isLoading, error, refetch, isFetching } = dealsApi.useGetDealsForFailureReviewQuery();

  const participantIds = useMemo(
    () => [...new Set(deals.flatMap((deal) => deal.participants))],
    [deals],
  );

  useEffect(() => {
    if (participantIds.length === 0) return;
    const subscriptions = participantIds.map((id) =>
      dispatch(usersApi.endpoints.getUserById.initiate(id)),
    );
    return () => subscriptions.forEach((subscription) => subscription.unsubscribe());
  }, [dispatch, participantIds]);

  const usersById = useAppSelector((state) =>
    participantIds.reduce<Record<string, User | undefined>>((acc, id) => {
      acc[id] = usersApi.endpoints.getUserById.select(id)(state).data;
      return acc;
    }, {}),
  );

  const getUserName = (id: string) => usersById[id]?.name?.trim() || "имя не указано";

  return (
    <>
      <Card variant="outlined" sx={{ borderRadius: 3 }}>
        <CardContent>
          <Box display="flex" justifyContent="space-between" alignItems="flex-start" gap={2} mb={2} flexWrap="wrap">
            <Box>
              <Typography variant="h5" fontWeight={700}>
                Очередь провалов сделок
              </Typography>
              <Typography variant="body2" color="text.secondary">
                Сделки, по которым достигнут порог голосов и требуется решение администратора.
              </Typography>
            </Box>
            <Stack direction="row" spacing={1} alignItems="center">
              <Chip color={deals.length > 0 ? "warning" : "success"} label={`${deals.length} в очереди`} />
              <Button variant="outlined" onClick={() => refetch()} disabled={isFetching}>
                Обновить
              </Button>
            </Stack>
          </Box>

          {isLoading ? (
            <Box display="flex" justifyContent="center" py={4}>
              <CircularProgress />
            </Box>
          ) : error ? (
            <Alert severity="error">
              {getStatusCode(error) === 403
                ? "Только администратор может просматривать очередь провалов"
                : "Не удалось загрузить очередь провалов"}
            </Alert>
          ) : deals.length === 0 ? (
            <Alert severity="success">Сделок, ожидающих решения по провалу, сейчас нет.</Alert>
          ) : (
            <Stack spacing={1.5}>
              {deals.map((deal) => (
                <Card
                  key={deal.id}
                  variant="outlined"
                  sx={{
                    borderRadius: 3,
                    borderColor: "warning.light",
                    background: "linear-gradient(180deg, rgba(255,152,0,0.06) 0%, rgba(255,152,0,0.015) 100%)",
                  }}
                >
                  <CardContent>
                    <Box display="flex" justifyContent="space-between" gap={2} flexWrap="wrap">
                      <Box>
                        <Typography variant="subtitle1" fontWeight={700}>
                          {deal.name?.trim() || "Сделка"}
                        </Typography>
                        <Typography variant="body2" color="text.secondary" mt={0.5}>
                          Участники: {deal.participants.map(getUserName).join(", ")}
                        </Typography>
                      </Box>
                      <Stack direction="row" spacing={1} flexWrap="wrap">
                        <Chip label={deal.status} color="warning" variant="outlined" />
                        <Button component={RouterLink} to={`/deals/${deal.id}`} variant="outlined" size="small">
                          Открыть сделку
                        </Button>
                      </Stack>
                    </Box>
                  </CardContent>
                </Card>
              ))}
            </Stack>
          )}
        </CardContent>
      </Card>

      {selectedDealId && (
        <FailureResolutionDialog dealId={selectedDealId} onClose={() => setSelectedDealId(null)} />
      )}
    </>
  );
}

export default FailureModerationQueue;
